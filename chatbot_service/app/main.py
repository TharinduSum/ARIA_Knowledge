import logging
import json
from contextlib import asynccontextmanager
from typing import List, Dict, Any

from fastapi import FastAPI, HTTPException, status, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import StreamingResponse, HTMLResponse
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates

import httpx
import markdown
from app.config import settings
from app.database import init_db, close_db, search_similar_chunks
from app.ollama_client import ollama_client
from app.schemas import (
    ChatRequest, 
    ChatResponse, 
    Citation, 
    HealthResponse,
    ChatMessage
)

# Set up logger
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger("chatbot.main")


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifespan context manager for database connection pool management."""
    logger.info("Initializing application resources...")
    await init_db()
    yield
    logger.info("Shutting down application resources...")
    await close_db()


app = FastAPI(
    title="ARIA Knowledge Chatbot Microservice",
    description="Asynchronous RAG Chatbot service connecting PostgreSQL (pgvector) and Ollama LLM.",
    version="1.0.0",
    lifespan=lifespan
)

# Enable CORS Middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Mount Static Files & Jinja2 Templates
app.mount("/static", StaticFiles(directory="app/static"), name="static")
templates = Jinja2Templates(directory="app/templates")


DEFAULT_SYSTEM_PROMPT = """You are an intelligent AI Knowledge Assistant powered by ARIA Knowledge Base.
Your job is to provide accurate, clear, and helpful answers strictly based on the provided reference context below.

RULES:
1. Use ONLY the information provided in the REFERENCE CONTEXT section.
2. If the context does not contain enough information to answer the question, state politely: "මගේ ලේඛනවල මේ පිළිබඳව තොරතුරු නොමැත." (or in English: "I could not find relevant information in the provided knowledge base documents.")
3. Do not invent facts or extrapolate beyond what is stated in the context.
4. Keep your answers concise, well-structured, and easy to read.
"""


def construct_rag_messages(
    user_query: str,
    chunks: List[Dict[str, Any]],
    history: List[ChatMessage],
    system_prompt_override: str = None
) -> List[Dict[str, str]]:
    """Build system message with context, history messages, and current user question."""
    system_instruction = system_prompt_override or DEFAULT_SYSTEM_PROMPT

    context_parts = []
    for idx, c in enumerate(chunks, 1):
        meta = c.get("metadata", {})
        source = meta.get("source") or meta.get("filename") or "Document"
        page = meta.get("page") or meta.get("page_number")
        
        loc_info = f"Source: {source}"
        if page:
            loc_info += f" (Page {page})"
            
        context_parts.append(f"--- [Chunk {idx} | {loc_info}] ---\n{c['content'].strip()}")

    formatted_context = "\n\n".join(context_parts) if context_parts else "No relevant document chunks found."

    full_system_content = f"{system_instruction}\n\n=== REFERENCE CONTEXT ===\n{formatted_context}\n========================="

    messages = [{"role": "system", "content": full_system_content}]

    if history:
        for msg in history[-10:]:
            messages.append({"role": msg.role, "content": msg.content})

    messages.append({"role": "user", "content": user_query})

    return messages


@app.get("/", response_class=HTMLResponse, tags=["UI"])
async def serve_ui(request: Request):
    """Serve web Chat UI using Jinja2 templates."""
    return templates.TemplateResponse(request=request, name="index.html")


@app.get("/health", response_model=HealthResponse, tags=["Health"])
async def health_check():
    """Health check endpoint to verify database and Ollama availability."""
    db_status = "healthy"
    ollama_status = "healthy"

    try:
        async with httpx.AsyncClient(timeout=5.0) as client:
            res = await client.get(f"{settings.ollama_url}/api/version")
            if res.status_code != 200:
                ollama_status = f"unhealthy (status {res.status_code})"
    except Exception as e:
        ollama_status = f"unreachable ({e})"

    return HealthResponse(
        status="healthy" if db_status == "healthy" and "unhealthy" not in ollama_status else "degraded",
        database=db_status,
        ollama=ollama_status
    )


def convert_markdown_to_html(text: str) -> str:
    """Convert Markdown string to clean HTML format."""
    if not text:
        return ""
    return markdown.markdown(
        text,
        extensions=[
            "extra",
            "codehilite",
            "nl2br",
            "tables",
            "sane_lists"
        ],
        extension_configs={
            "codehilite": {
                "css_class": "highlight",
                "guess_lang": False,
            }
        }
    )


@app.post("/api/v1/chat", response_model=ChatResponse, tags=["Chat"])
async def chat_endpoint(request: ChatRequest):
    """
    Standard non-streaming Chat endpoint.
    1. Embeds user query via Ollama.
    2. Vector similarity search in PostgreSQL (pgvector).
    3. Generates LLM response via Ollama.
    4. Converts response from Markdown to HTML.
    """
    try:
        top_k = request.top_k or settings.default_top_k
        logger.info(f"Received chat request: {request.message[:50]}... (top_k={top_k})")

        query_embedding = await ollama_client.get_embedding(request.message)
        chunks = await search_similar_chunks(query_embedding, top_k=top_k)

        messages = construct_rag_messages(
            user_query=request.message,
            chunks=chunks,
            history=request.history,
            system_prompt_override=request.system_prompt
        )

        answer = await ollama_client.chat(messages)
        html_message = convert_markdown_to_html(answer)

        citations = [
            Citation(
                content=c["content"],
                metadata=c["metadata"],
                similarity=c["similarity"]
            )
            for c in chunks
        ]

        return ChatResponse(
            message=answer,
            html_message=html_message,
            citations=citations,
            model=settings.ollama_llm_model
        )

    except Exception as e:
        logger.error(f"Error handling chat request: {e}", exc_info=True)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Chat generation failed: {str(e)}"
        )


@app.post("/api/v1/chat/stream", tags=["Chat"])
async def chat_stream_endpoint(request: ChatRequest):
    """
    Real-time streaming Chat endpoint using Server-Sent Events (SSE).
    """
    try:
        top_k = request.top_k or settings.default_top_k
        logger.info(f"Received streaming chat request: {request.message[:50]}...")

        query_embedding = await ollama_client.get_embedding(request.message)
        chunks = await search_similar_chunks(query_embedding, top_k=top_k)

        messages = construct_rag_messages(
            user_query=request.message,
            chunks=chunks,
            history=request.history,
            system_prompt_override=request.system_prompt
        )

        citations_data = [
            {
                "content": c["content"],
                "metadata": c["metadata"],
                "similarity": c["similarity"]
            }
            for c in chunks
        ]

        async def sse_event_generator():
            meta_event = json.dumps({"type": "citations", "citations": citations_data})
            yield f"data: {meta_event}\n\n"

            async for token_event in ollama_client.stream_chat(messages):
                yield token_event

        return StreamingResponse(
            sse_event_generator(),
            media_type="text/event-stream"
        )

    except Exception as e:
        logger.error(f"Error handling streaming chat request: {e}", exc_info=True)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Streaming chat generation failed: {str(e)}"
        )


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("app.main:app", host=settings.host, port=settings.port, reload=True)
