# ARIA Knowledge Base

A high-performance document processing and semantic search system built with **Golang**, **LangChain for Go (langchaingo)**, **pgvector**, and **React**. It allows users to upload PDF documents, extract text, split it into chunks, generate embeddings via Ollama, and perform similarity search with an interactive, glassmorphic UI.

---

## ✨ Features

- **Golang Backend**: Fast and robust REST API utilizing standard `net/http` routing.
- **LangChain for Go**: Built-in PDF loading (`documentloaders.NewPDF`), recursive character text splitting (`textsplitter.NewRecursiveCharacter`), and Ollama embedding integration.
- **pgvector Integration**: Stores vector embeddings and metadata directly in PostgreSQL, performing cosine similarity queries.
- **Glassmorphic React UI**: Modern dark theme dashboard with drag-and-drop file upload, real-time ingestion status, settings config, and semantic search interface.
- **Full Dockerized Setup**: Spin up PostgreSQL (with pgvector), the Go Backend, and the React Frontend using a single command.

---

## 🛠️ Tech Stack

- **Backend**: Golang 1.25, `github.com/tmc/langchaingo` (embeddings, vector store, document loaders)
- **Frontend**: React + Vite, Vanilla CSS, Lucide Icons
- **Database**: PostgreSQL 16 + pgvector extension
- **Embedding Provider**: Ollama (fully local model execution)

---

## 🚀 Getting Started

### Prerequisites

1. **Ollama Server (`192.168.10.10`)**: Must be running Ollama and have the embedding model pulled:
   ```bash
   ollama pull nomic-embed-text
   ```
2. **Database Server (`192.168.10.20`)**: Must have **PostgreSQL** running with the **pgvector** extension active. Create a database named `aria_knowledge` and enable the extension:
   ```sql
   CREATE DATABASE aria_knowledge;
   \c aria_knowledge;
   CREATE EXTENSION IF NOT EXISTS vector;
   ```
3. **Docker Engine**: The server `192.168.10.20` must have Docker and Docker Compose installed to run the backend and frontend containers.

### Running the Application

1. Connect to your database server `192.168.10.20` via SSH and clone this repository.
2. Spin up the application containers from the project root:
   ```bash
   docker compose up --build
   ```
3. Open your browser and navigate to:
   - **Frontend UI**: `http://192.168.10.20:5173`
   - **Go Backend API**: `http://192.168.10.20:8080/api/health`

### 💻 Running Locally (Without Docker)

If you wish to run the backend and frontend directly on your local machine for development:

#### 1. Run Go Backend

From the project root directory:

```bash
cd backend
# Edit settings in the .env file (pre-configured for you)
go run cmd/server/main.go
```

The server will run on `http://localhost:8080`.

#### 2. Run React Frontend

Open a new terminal window at the project root directory:

```bash
cd frontend
npm install
npm run dev
```

The client will run on `http://localhost:5173`. The UI automatically binds to the local backend.

---

## 📂 Project Structure

- `backend/`: Go source code (handlers, config, services, and Dockerfile)
- `frontend/`: React + Vite client (App component, custom styles, and Dockerfile)
- `docker-compose.yml`: Multi-container Orchestration (Postgres + Backend + Frontend)
- `README.md`: Setup and documentation

---

## 🌐 API Reference

### 1. Ingestion (`POST /api/upload`)

Accepts a PDF file, parses it, creates overlap chunks, and saves vector embeddings to PostgreSQL.

- **Content-Type**: `multipart/form-data`
- **Form Values**:
  - `file`: PDF binary file
  - `chunk_size`: Maximum characters per chunk (default `1000`)
  - `chunk_overlap`: Overlap size between chunks (default `150`)

### 2. Similarity Search (`POST /api/search`)

Performs a cosine similarity search across all indexed chunks.

- **Content-Type**: `application/json`
- **Request Body**:
  ```json
  {
    "query": "What are the rules for hybrid work?",
    "limit": 5
  }
  ```
- **Response Body**:
  ```json
  {
    "query": "What are the rules for hybrid work?",
    "results": [
      {
        "content": "...employees can work up to 2 days remotely...",
        "score": 0.892,
        "metadata": {
          "source": "handbook.pdf",
          "page": 4
        }
      }
    ]
  }
  ```
  > > > > > > > 42bab4a (initial commit)
