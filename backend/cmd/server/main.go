package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/aria-knowledge/backend/internal/config"
	"github.com/aria-knowledge/backend/internal/handlers"
	"github.com/aria-knowledge/backend/internal/ingest"
	"github.com/aria-knowledge/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := config.Load()

	log.Printf("Starting ARIA Knowledge Backend on port %s...", cfg.Port)

	// Context for initialization
	ctx := context.Background()

	var err error

	// Initialize database pool
	var dbPool *pgxpool.Pool
	log.Println("Connecting to Postgres database pool (retrying up to 5 times)...")
	for i := 1; i <= 5; i++ {
		dbPool, err = pgxpool.New(ctx, cfg.DatabaseURL)
		if err == nil {
			err = dbPool.Ping(ctx)
			if err == nil {
				break
			}
		}
		log.Printf("[Attempt %d/5] Failed to connect to DB pool: %v. Retrying in 5 seconds...", i, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatalf("Fatal: failed to connect to database pool: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection pool established!")

	// Retry connecting to Ollama
	var vectorService *service.VectorService

	log.Println("Initializing services (retrying up to 5 times for database and Ollama ready)...")
	for i := 1; i <= 5; i++ {
		vectorService, err = service.NewVectorService(ctx, cfg)
		if err == nil {
			break
		}
		log.Printf("[Attempt %d/5] Failed to initialize Vector Service: %v. Retrying in 5 seconds...", i, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatalf("Fatal: failed to initialize Vector Service after retries: %v", err)
	}
	log.Println("Vector Service successfully initialized!")

	// Initialize JobService
	jobService := ingest.NewJobService(dbPool, cfg.OllamaURL, cfg.OllamaVisionModel, vectorService)
	log.Println("Initializing jobs database schema...")
	if err := jobService.InitSchema(ctx); err != nil {
		log.Fatalf("Fatal: failed to initialize jobs schema: %v", err)
	}

	// Start async worker pool
	jobService.StartWorkers(2)
	defer jobService.StopWorkers()

	uploadHandler := handlers.NewUploadHandler(jobService)
	jobsHandler := handlers.NewJobsHandler(jobService)
	queryHandler := handlers.NewQueryHandler(vectorService)

	// Setup simple router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Upload, Search, and Jobs endpoints
	mux.Handle("/api/upload", uploadHandler)
	mux.Handle("/api/search", queryHandler)
	mux.Handle("/api/jobs", jobsHandler)

	// Swagger documentation endpoints
	mux.HandleFunc("/swagger/doc.json", handlers.SwaggerJSONHandler)
	mux.HandleFunc("/swagger/", handlers.SwaggerUIHandler)
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})

	// Wrap in CORS middleware
	handler := corsMiddleware(mux)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	log.Printf("Server listening on http://localhost:%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server listen failed: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
