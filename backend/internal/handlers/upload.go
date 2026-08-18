package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/aria-knowledge/backend/internal/service"
)

type UploadHandler struct {
	pdfService    *service.PDFService
	vectorService *service.VectorService
}

func NewUploadHandler(pdfService *service.PDFService, vectorService *service.VectorService) *UploadHandler {
	return &UploadHandler{
		pdfService:    pdfService,
		vectorService: vectorService,
	}
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle OPTIONS preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Max 50 MB uploads
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Failed to retrieve file from form: %v", err)
		http.Error(w, "Failed to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Parse settings
	chunkSizeStr := r.FormValue("chunk_size")
	chunkOverlapStr := r.FormValue("chunk_overlap")

	chunkSize := 1000
	if val, err := strconv.Atoi(chunkSizeStr); err == nil && val > 0 {
		chunkSize = val
	}

	chunkOverlap := 150
	if val, err := strconv.Atoi(chunkOverlapStr); err == nil && val >= 0 {
		chunkOverlap = val
	}

	log.Printf("Uploading file: %s (size: %d bytes), ChunkSize: %d, ChunkOverlap: %d", header.Filename, header.Size, chunkSize, chunkOverlap)

	// Process the PDF into vectorstore documents
	docs, err := h.pdfService.ProcessPDF(r.Context(), file, header.Size, chunkSize, chunkOverlap, header.Filename)
	if err != nil {
		log.Printf("Error processing PDF %s: %v", header.Filename, err)
		http.Error(w, "Error processing PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Index in pgvector
	err = h.vectorService.AddDocuments(r.Context(), docs)
	if err != nil {
		log.Printf("Error adding documents to vector store: %v", err)
		http.Error(w, "Error saving vectors: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"status":       "success",
		"message":      "File processed and indexed successfully",
		"chunks_count": len(docs),
		"filename":     header.Filename,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
