package ingest

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ledongthuc/pdf"
)

// ProcessDocument processes the PDF file page-by-page.
// For each page, it classifies the page, runs either the digital or scanned extraction path,
// captures captions for any images, merges the text, and returns the merged pages.
func ProcessDocument(ctx context.Context, pdfPath string, ollamaURL string, visionModel string) ([]MergedPage, error) {
	// 1. Open the PDF using ledongthuc/pdf to get total pages
	f, r, err := pdf.Open(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF %q: %w", pdfPath, err)
	}
	defer f.Close()

	totalPage := r.NumPage()
	log.Printf("Document %s has %d pages. Starting ingestion pipeline...", pdfPath, totalPage)

	var mergedPages []MergedPage

	// 2. Iterate through pages (1-indexed)
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		log.Printf("Processing page %d/%d...", pageIndex, totalPage)

		p := r.Page(pageIndex)
		if p.V.IsNull() {
			log.Printf("Page %d is null, skipping", pageIndex)
			continue
		}

		// Step 1: Classify
		isScanned, _, err := IsScannedPage(p, DefaultThreshold)
		if err != nil {
			log.Printf("Error classifying page %d: %v. Defaulting to scanned path", pageIndex, err)
			isScanned = true
		}

		var pageText string
		var imagePaths []string
		var sourceType string

		if !isScanned {
			// Step 2: Digital text path
			log.Printf("Page %d classified as DIGITAL", pageIndex)
			sourceType = "digital"
			pageText, imagePaths, err = ExtractDigitalPage(pdfPath, pageIndex)
			if err != nil {
				log.Printf("Error extracting digital page %d: %v. Falling back to scanned path", pageIndex, err)
				// Fallback to scanned
				isScanned = true
			}
		}

		// Re-evaluate if we fell back or if it was classified as scanned initially
		if isScanned {
			// Step 3: Scanned path
			log.Printf("Page %d classified as SCANNED", pageIndex)
			sourceType = "scanned"
			var rasterImgPath string
			pageText, rasterImgPath, err = ExtractScannedPage(pdfPath, pageIndex)
			if err != nil {
				return nil, fmt.Errorf("failed to extract scanned page %d: %w", pageIndex, err)
			}
			if rasterImgPath != "" {
				imagePaths = []string{rasterImgPath}
			}
		}

		// Step 4: Captioning (multimodal LLM description)
		var captions []string
		if len(imagePaths) > 0 {
			for _, imgPath := range imagePaths {
				log.Printf("Captioning image: %s ...", imgPath)
				caption, err := CaptionImage(ctx, ollamaURL, visionModel, imgPath)
				if err != nil {
					log.Printf("Warning: failed to caption image %s: %v", imgPath, err)
					// Don't fail the whole page if a single caption fails, just continue
					continue
				}
				if caption != "" {
					captions = append(captions, caption)
				}
			}

			// Cleanup image files immediately after captioning
			for _, imgPath := range imagePaths {
				_ = os.Remove(imgPath)
				// Remove the parent temp directory as well if it's empty
				parentDir := filepath.Dir(imgPath)
				_ = os.Remove(parentDir) // Will only succeed if empty
			}
		} else if isScanned && pageText == "" {
			// In case page is scanned but no image path returned (unlikely)
			log.Printf("Scanned page %d has no image path", pageIndex)
		}

		// Step 5: Merge base text with captions
		merged := MergePageContent(pageIndex, sourceType, pageText, captions)
		mergedPages = append(mergedPages, merged)
	}

	return mergedPages, nil
}
