package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/aria-knowledge/backend/internal/config"
	"github.com/aria-knowledge/backend/internal/handlers"
	"github.com/aria-knowledge/backend/internal/service"
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

	// Retry connecting to Postgres and Ollama
	var vectorService *service.VectorService
	var err error

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

	pdfService := service.NewPDFService()

	uploadHandler := handlers.NewUploadHandler(pdfService, vectorService)
	queryHandler := handlers.NewQueryHandler(vectorService)

	// Setup simple router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Upload & Search endpoints
	mux.Handle("/api/upload", uploadHandler)
	mux.Handle("/api/search", queryHandler)

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
		ReadTimeout:  120 * time.Second, // PDF processing can take a while
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
