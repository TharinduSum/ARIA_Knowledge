package ingest

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// CaptionImage sends the image file to the multimodal Ollama model and gets a text description.
func CaptionImage(ctx context.Context, ollamaURL string, modelName string, imagePath string) (string, error) {
	// 1. Read the image file bytes
	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file %q: %w", imagePath, err)
	}

	// 2. Initialize Ollama client for the vision model
	llm, err := ollama.New(
		ollama.WithServerURL(ollamaURL),
		ollama.WithModel(modelName),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create Ollama client for vision model: %w", err)
	}

	// 3. Determine MIME type based on extension
	mimeType := "image/png"
	lowerPath := strings.ToLower(imagePath)
	if strings.HasSuffix(lowerPath, ".jpg") || strings.HasSuffix(lowerPath, ".jpeg") {
		mimeType = "image/jpeg"
	} else if strings.HasSuffix(lowerPath, ".gif") {
		mimeType = "image/gif"
	} else if strings.HasSuffix(lowerPath, ".bmp") {
		mimeType = "image/bmp"
	} else if strings.HasSuffix(lowerPath, ".webp") {
		mimeType = "image/webp"
	}

	// 4. Construct the prompt asking Ollama to describe the chart/figure and its data
	prompt := "Describe this chart, figure, or image. Identify any data, labels, trends, or visual structures shown."

	resp, err := llm.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.BinaryPart(mimeType, imageBytes),
				llms.TextPart(prompt),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate caption from Ollama: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from Ollama")
	}

	return resp.Choices[0].Content, nil
}
