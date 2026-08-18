package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	OllamaURL   string
	OllamaModel string
	Port        string
}

func Load() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/aria_knowledge?sslmode=disable"
	}

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "nomic-embed-text"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DatabaseURL: dbURL,
		OllamaURL:   ollamaURL,
		OllamaModel: ollamaModel,
		Port:        port,
	}
}
