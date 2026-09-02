import json
import logging
import httpx
from typing import List, Dict, Any, AsyncGenerator
from app.config import settings

logger = logging.getLogger("chatbot.ollama")


class OllamaClient:
    def __init__(self, base_url: str = None):
        self.base_url = (base_url or settings.ollama_url).rstrip("/")

    async def get_embedding(self, text: str, model: str = None) -> List[float]:
        """Fetch vector embedding for text using Ollama embed endpoint."""
        embed_model = model or settings.ollama_embed_model
        url = f"{self.base_url}/api/embeddings"
        
        payload = {
            "model": embed_model,
            "prompt": text,
        }

        async with httpx.AsyncClient(timeout=30.0) as client:
            try:
                response = await client.post(url, json=payload)
                response.raise_for_status()
                data = response.json()
                
                # Check for standard 'embedding' field or 'embeddings' array
                if "embedding" in data:
                    return data["embedding"]
                elif "embeddings" in data and len(data["embeddings"]) > 0:
                    return data["embeddings"][0]
                else:
                    raise ValueError(f"Unexpected embedding response structure from Ollama: {data}")
            except Exception as e:
                logger.error(f"Error fetching embedding from Ollama ({url}): {e}")
                raise RuntimeError(f"Ollama embedding request failed: {e}") from e

    async def chat(
        self, 
        messages: List[Dict[str, str]], 
        model: str = None
    ) -> str:
        """Send chat messages payload to Ollama /api/chat endpoint (non-streaming)."""
        llm_model = model or settings.ollama_llm_model
        url = f"{self.base_url}/api/chat"

        payload = {
            "model": llm_model,
            "messages": messages,
            "stream": False,
        }

        async with httpx.AsyncClient(timeout=120.0) as client:
            try:
                response = await client.post(url, json=payload)
                response.raise_for_status()
                data = response.json()
                return data.get("message", {}).get("content", "")
            except Exception as e:
                logger.error(f"Error generating chat completion from Ollama: {e}")
                raise RuntimeError(f"Ollama chat completion failed: {e}") from e

    async def stream_chat(
        self, 
        messages: List[Dict[str, str]], 
        model: str = None
    ) -> AsyncGenerator[str, None]:
        """Async generator streaming Server-Sent Events (SSE) from Ollama /api/chat."""
        llm_model = model or settings.ollama_llm_model
        url = f"{self.base_url}/api/chat"

        payload = {
            "model": llm_model,
            "messages": messages,
            "stream": True,
        }

        async with httpx.AsyncClient(timeout=180.0) as client:
            try:
                async with client.stream("POST", url, json=payload) as response:
                    response.raise_for_status()
                    async for line in response.aiter_lines():
                        if line:
                            try:
                                chunk = json.loads(line)
                                token = chunk.get("message", {}).get("content", "")
                                is_done = chunk.get("done", False)
                                
                                sse_data = json.dumps({"token": token, "done": is_done})
                                yield f"data: {sse_data}\n\n"
                                
                                if is_done:
                                    break
                            except json.JSONDecodeError:
                                continue
            except Exception as e:
                logger.error(f"Streaming error from Ollama: {e}")
                err_data = json.dumps({"error": str(e), "done": True})
                yield f"data: {err_data}\n\n"


ollama_client = OllamaClient()
