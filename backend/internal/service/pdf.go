package service

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type PDFService struct{}

func NewPDFService() *PDFService {
	return &PDFService{}
}

func (s *PDFService) ProcessPDF(ctx context.Context, r io.Reader, size int64, chunkSize, chunkOverlap int, filename string) ([]schema.Document, error) {
	// Create a temp file to read from
	tempFile, err := os.CreateTemp("", "aria-pdf-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, r); err != nil {
		return nil, fmt.Errorf("failed to write uploaded file to temp path: %w", err)
	}

	// Seek to beginning
	if _, err := tempFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek temp file: %w", err)
	}

	// Load the document using documentloaders
	loader := documentloaders.NewPDF(tempFile, size)
	docs, err := loader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PDF: %w", err)
	}

	// Inject metadata like filename and page number into documents
	for i := range docs {
		if docs[i].Metadata == nil {
			docs[i].Metadata = make(map[string]any)
		}
		docs[i].Metadata["source"] = filename
		// If page number is not already present, we can add it (1-based index)
		if _, ok := docs[i].Metadata["page"]; !ok {
			docs[i].Metadata["page"] = i + 1
		}
	}

	// Initialize the splitter
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(chunkSize),
		textsplitter.WithChunkOverlap(chunkOverlap),
	)

	// Split documents into chunks
	splitDocs, err := textsplitter.SplitDocuments(splitter, docs)
	if err != nil {
		return nil, fmt.Errorf("failed to split documents: %w", err)
	}

	return splitDocs, nil
}
