package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/aria-knowledge/backend/internal/service"
)

type QueryRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type QueryHandler struct {
	vectorService *service.VectorService
}

func NewQueryHandler(vectorService *service.VectorService) *QueryHandler {
	return &QueryHandler{
		vectorService: vectorService,
	}
}

func (h *QueryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle OPTIONS preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req QueryRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Failed to decode search request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	if req.Limit <= 0 {
		req.Limit = 5
	}

	results, err := h.vectorService.Search(r.Context(), req.Query, req.Limit)
	if err != nil {
		log.Printf("Search failed for query %q: %v", req.Query, err)
		http.Error(w, "Search execution failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"query":   req.Query,
		"results": results,
	})
}
