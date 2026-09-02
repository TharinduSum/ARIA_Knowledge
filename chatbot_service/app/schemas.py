from typing import List, Optional, Any, Dict
from pydantic import BaseModel, Field


class ChatMessage(BaseModel):
    role: str = Field(..., description="Role of the message sender: 'user', 'assistant', or 'system'")
    content: str = Field(..., description="Content text of the message")


class ChatRequest(BaseModel):
    message: str = Field(..., min_length=1, description="Current user query")
    history: Optional[List[ChatMessage]] = Field(default=[], description="Optional conversation history")
    top_k: Optional[int] = Field(default=5, ge=1, le=20, description="Number of context chunks to retrieve")
    stream: bool = Field(default=False, description="Whether to stream the LLM response via SSE")
    system_prompt: Optional[str] = Field(
        default=None, 
        description="Optional custom system prompt override"
    )


class Citation(BaseModel):
    content: str = Field(..., description="Retrieved chunk text content")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Metadata such as source filename, page, etc.")
    similarity: float = Field(..., description="Cosine similarity score (0.0 to 1.0)")


class ChatResponse(BaseModel):
    message: str = Field(..., description="Generated assistant response in Markdown format")
    html_message: Optional[str] = Field(default=None, description="HTML converted from Markdown response")
    citations: List[Citation] = Field(default_factory=list, description="Source context chunks used")
    model: str = Field(..., description="LLM model used for inference")


class HealthResponse(BaseModel):
    status: str
    database: str
    ollama: str
