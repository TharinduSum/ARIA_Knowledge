# ARIA Knowledge - FastAPI Chatbot Microservice

A production-ready, asynchronous FastAPI microservice for RAG (Retrieval-Augmented Generation) Chatbot connecting to the PostgreSQL (`aria_knowledge` + `pgvector`) database and Ollama AI engine.

## 🚀 Features

- **FastAPI Async Engine**: High-performance HTTP server built with Pydantic v2 and `asyncpg`.
- **pgvector Cosine Search**: Direct vector similarity search using PostgreSQL `<=>` operator.
- **Ollama Integration**: Async integration with Ollama embeddings (`nomic-embed-text`) and LLM text generation (`llama3.2` / `qwen2.5`).
- **Real-time SSE Streaming**: `/api/v1/chat/stream` supporting token streaming for interactive chat UIs.
- **CORS Enabled**: Ready for frontend UI connection.

---

## 🛠️ Quick Start

### 1. Installation

```bash
cd chatbot_service
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### 2. Environment Configuration

Copy `.env.example` to `.env` and adjust the variables if needed:

```bash
cp .env.example .env
```

Default settings:
- `DATABASE_URL`: `postgresql://postgres:postgres@192.168.10.20:5432/aria_knowledge`
- `OLLAMA_URL`: `http://192.168.10.10:11434`
- `OLLAMA_EMBED_MODEL`: `nomic-embed-text`
- `OLLAMA_LLM_MODEL`: `llama3.2`

### 3. Run the Server

```bash
uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
```

---

## 🌐 API Reference

### 1. Health Check (`GET /health`)
```bash
curl -X GET http://localhost:8000/health
```

### 2. Standard Chat (`POST /api/v1/chat`)
```bash
curl -X POST http://localhost:8000/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What are the rules for remote work?",
    "top_k": 5
  }'
```

### 3. Streaming Chat (`POST /api/v1/chat/stream`)
```bash
curl -X POST http://localhost:8000/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What is the policy?",
    "top_k": 3
  }'
```
