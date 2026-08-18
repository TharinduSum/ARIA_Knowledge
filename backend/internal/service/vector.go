package service

import (
	"context"
	"fmt"
	"log"

	"github.com/aria-knowledge/backend/internal/config"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

type VectorService struct {
	Store vectorstores.VectorStore
}

func NewVectorService(ctx context.Context, cfg *config.Config) (*VectorService, error) {
	log.Printf("Connecting to Ollama at %s with model %s...", cfg.OllamaURL, cfg.OllamaModel)
	llm, err := ollama.New(
		ollama.WithServerURL(cfg.OllamaURL),
		ollama.WithModel(cfg.OllamaModel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}

	log.Println("Creating embedder client...")
	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	log.Printf("Initializing pgvector store with DB URL %s...", cfg.DatabaseURL)
	store, err := pgvector.New(
		ctx,
		pgvector.WithConnectionURL(cfg.DatabaseURL),
		pgvector.WithEmbedder(embedder),
		// Using a specific collection table name
		pgvector.WithCollectionName("aria_knowledge"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pgvector store: %w", err)
	}

	return &VectorService{Store: store}, nil
}

func (s *VectorService) AddDocuments(ctx context.Context, docs []schema.Document) error {
	log.Printf("Adding %d documents to vector store...", len(docs))
	_, err := s.Store.AddDocuments(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}
	return nil
}

type SearchResult struct {
	Content  string         `json:"content"`
	Score    float32        `json:"score"`
	Metadata map[string]any `json:"metadata"`
}

func (s *VectorService) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	log.Printf("Searching for query: %q (limit: %d)...", query, limit)
	
	// SimilaritySearch returns documents matching the query.
	// Many vectorstores return cosine similarity score as part of options.
	// In langchaingo, we can use SimilaritySearch with score options.
	docs, err := s.Store.SimilaritySearch(
		ctx,
		query,
		limit,
		// pgvector has standard similarity thresholds if supported, or we can just run normal search.
	)
	if err != nil {
		return nil, fmt.Errorf("failed similarity search: %w", err)
	}

	results := make([]SearchResult, len(docs))
	for i, doc := range docs {
		// Note: Score is typically passed inside the metadata or via option.
		// If score is not explicitly returned in doc, score might be 0.
		// Let's inspect the document.
		scoreVal := float32(0.0)
		if score, ok := doc.Metadata["score"].(float32); ok {
			scoreVal = score
		} else if score, ok := doc.Metadata["score"].(float64); ok {
			scoreVal = float32(score)
		}
		
		results[i] = SearchResult{
			Content:  doc.PageContent,
			Score:    scoreVal,
			Metadata: doc.Metadata,
		}
	}

	return results, nil
}
