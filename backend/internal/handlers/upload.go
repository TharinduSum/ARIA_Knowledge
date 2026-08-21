package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/aria-knowledge/backend/internal/ingest"
	"github.com/google/uuid"
)

type UploadHandler struct {
	jobService *ingest.JobService
}

func NewUploadHandler(jobService *ingest.JobService) *UploadHandler {
	return &UploadHandler{
		jobService: jobService,
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

	log.Printf("Uploading file: %s (size: %d bytes) for async ingestion", header.Filename, header.Size)

	// Create and queue the ingestion job
	jobID, err := h.jobService.CreateJob(r.Context(), header.Filename, file)
	if err != nil {
		log.Printf("Error creating ingestion job for %s: %v", header.Filename, err)
		http.Error(w, "Error queuing PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"status":   "queued",
		"message":  "File queued for async ingestion and processing",
		"job_id":   jobID.String(),
		"filename": header.Filename,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

type JobsHandler struct {
	jobService *ingest.JobService
}

func NewJobsHandler(jobService *ingest.JobService) *JobsHandler {
	return &JobsHandler{
		jobService: jobService,
	}
}

func (h *JobsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle OPTIONS preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobIDStr := r.URL.Query().Get("id")
	if jobIDStr == "" {
		http.Error(w, "Missing job id parameter", http.StatusBadRequest)
		return
	}

	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		http.Error(w, "Invalid job id format", http.StatusBadRequest)
		return
	}

	job, err := h.jobService.GetJobStatus(r.Context(), jobID)
	if err != nil {
		log.Printf("Error retrieving job status for %s: %v", jobID, err)
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}
