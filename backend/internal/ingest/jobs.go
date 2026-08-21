package ingest

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type JobStatus string

const (
	StatusQueued     JobStatus = "queued"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

type IngestionJob struct {
	ID           uuid.UUID
	Filename     string
	Status       JobStatus
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// VectorStoreAdder abstracts VectorService to allow decouple indexing
type VectorStoreAdder interface {
	AddDocuments(ctx context.Context, docs []schema.Document) error
}

type JobService struct {
	db          *pgxpool.Pool
	ollamaURL   string
	visionModel string
	vectorStore VectorStoreAdder
	jobQueue    chan uuid.UUID
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewJobService(db *pgxpool.Pool, ollamaURL, visionModel string, vectorStore VectorStoreAdder) *JobService {
	ctx, cancel := context.WithCancel(context.Background())
	return &JobService{
		db:          db,
		ollamaURL:   ollamaURL,
		visionModel: visionModel,
		vectorStore: vectorStore,
		jobQueue:    make(chan uuid.UUID, 100),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// InitSchema creates the necessary database table if it doesn't exist.
func (s *JobService) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS ingestion_jobs (
		id UUID PRIMARY KEY,
		filename VARCHAR(255) NOT NULL,
		status VARCHAR(50) NOT NULL,
		error_message TEXT,
		file_data BYTEA,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := s.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create ingestion_jobs table: %w", err)
	}
	return nil
}

// StartWorkers spawns the worker pool.
func (s *JobService) StartWorkers(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
	log.Printf("Started %d async ingestion workers", numWorkers)
}

// StopWorkers shuts down the worker pool gracefully.
func (s *JobService) StopWorkers() {
	s.cancel()
	close(s.jobQueue)
	s.wg.Wait()
	log.Println("Stopped all async ingestion workers")
}

// CreateJob registers a new job in the database, stores the file bytes, and queues it.
func (s *JobService) CreateJob(ctx context.Context, filename string, r io.Reader) (uuid.UUID, error) {
	jobID := uuid.New()

	// Read file bytes
	fileData, err := io.ReadAll(r)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to read file data: %w", err)
	}

	query := `
	INSERT INTO ingestion_jobs (id, filename, status, file_data)
	VALUES ($1, $2, $3, $4)
	`
	_, err = s.db.Exec(ctx, query, jobID, filename, string(StatusQueued), fileData)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to insert job into database: %w", err)
	}

	// Queue job
	s.jobQueue <- jobID
	log.Printf("Job %s queued for filename %s", jobID, filename)

	return jobID, nil
}

// GetJobStatus retrieves the current state of a job.
func (s *JobService) GetJobStatus(ctx context.Context, id uuid.UUID) (*IngestionJob, error) {
	query := `
	SELECT id, filename, status, COALESCE(error_message, ''), created_at, updated_at
	FROM ingestion_jobs
	WHERE id = $1
	`
	var job IngestionJob
	err := s.db.QueryRow(ctx, query, id).Scan(&job.ID, &job.Filename, &job.Status, &job.ErrorMessage, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job status: %w", err)
	}
	return &job, nil
}

// worker loops on the job queue.
func (s *JobService) worker(id int) {
	defer s.wg.Done()
	log.Printf("Worker %d ready", id)

	for {
		select {
		case <-s.ctx.Done():
			return
		case jobID, ok := <-s.jobQueue:
			if !ok {
				return
			}
			log.Printf("[Worker %d] Processing job %s...", id, jobID)
			err := s.processJob(s.ctx, jobID)
			if err != nil {
				log.Printf("[Worker %d] Job %s failed: %v", id, jobID, err)
				s.updateJobStatus(s.ctx, jobID, StatusFailed, err.Error())
			} else {
				log.Printf("[Worker %d] Job %s completed successfully", id, jobID)
				s.updateJobStatus(s.ctx, jobID, StatusCompleted, "")
			}
		}
	}
}

func (s *JobService) updateJobStatus(ctx context.Context, id uuid.UUID, status JobStatus, errMsg string) {
	query := `
	UPDATE ingestion_jobs
	SET status = $1, error_message = $2, updated_at = CURRENT_TIMESTAMP
	WHERE id = $3
	`
	_, err := s.db.Exec(ctx, query, string(status), errMsg, id)
	if err != nil {
		log.Printf("Failed to update status for job %s: %v", id, err)
	}
}

func (s *JobService) processJob(ctx context.Context, jobID uuid.UUID) error {
	// 1. Mark as processing
	s.updateJobStatus(ctx, jobID, StatusProcessing, "")

	var filename string
	var fileData []byte
	query := `SELECT filename, file_data FROM ingestion_jobs WHERE id = $1`
	err := s.db.QueryRow(ctx, query, jobID).Scan(&filename, &fileData)
	if err != nil {
		return fmt.Errorf("failed to fetch job data from db: %w", err)
	}

	// 2. Write bytes to a temporary PDF file
	tempFile, err := os.CreateTemp("", "aria-job-*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := tempFile.Write(fileData); err != nil {
		return fmt.Errorf("failed to write temp PDF: %w", err)
	}

	// 3. Run processing pipeline
	mergedPages, err := ProcessDocument(ctx, tempFile.Name(), s.ollamaURL, s.visionModel)
	if err != nil {
		return fmt.Errorf("pipeline processing failed: %w", err)
	}

	// 4. Chunk and load to vector store
	var docs []schema.Document
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(1000),   // Standard chunk size
		textsplitter.WithChunkOverlap(150), // Standard chunk overlap
	)

	for _, page := range mergedPages {
		// Use langchaingo to split this page's text
		pageDocs, err := textsplitter.CreateDocuments(splitter, []string{page.MergedText}, []map[string]any{
			{
				"source":      filename,
				"page":        page.PageNum,
				"source_type": page.SourceType,
				"has_images":  page.HasImages,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to split page %d: %w", page.PageNum, err)
		}
		docs = append(docs, pageDocs...)
	}

	// 5. Add documents to vector store
	if len(docs) > 0 {
		if err := s.vectorStore.AddDocuments(ctx, docs); err != nil {
			return fmt.Errorf("failed to store vectors: %w", err)
		}
	}

	return nil
}
