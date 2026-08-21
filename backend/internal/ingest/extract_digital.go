package ingest

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// ExtractDigitalPage extracts text using ledongthuc/pdf and saves embedded images using pdfcpu.
// It returns the page text, a slice of local image file paths, and any error.
// The caller is responsible for deleting the returned image files and directories when done.
func ExtractDigitalPage(pdfPath string, pageNum int) (string, []string, error) {
	// 1. Open the PDF to extract text via ledongthuc/pdf
	f, r, err := pdf.Open(pdfPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to open PDF for text extraction: %w", err)
	}
	defer f.Close()

	if pageNum < 1 || pageNum > r.NumPage() {
		return "", nil, fmt.Errorf("page number %d out of range (1-%d)", pageNum, r.NumPage())
	}

	page := r.Page(pageNum)
	if page.V.IsNull() {
		return "", nil, nil
	}

	rows, err := page.GetTextByRow()
	var text string
	if err == nil {
		var buf bytes.Buffer
		for _, row := range rows {
			for _, t := range row.Content {
				buf.WriteString(t.S)
				buf.WriteString(" ")
			}
			buf.WriteString("\n")
		}
		text = buf.String()
	}

	// 2. Create a temporary directory to save extracted images
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("aria-images-page-%d-*", pageNum))
	if err != nil {
		return text, nil, fmt.Errorf("failed to create temp dir for images: %w", err)
	}

	// 3. Extract images using pdfcpu/pkg/api
	selectedPages := []string{strconv.Itoa(pageNum)}
	err = api.ExtractImagesFile(pdfPath, tempDir, selectedPages, nil)
	if err != nil {
		// If error is related to no images, we can ignore it and return empty image slice.
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "no images") || strings.Contains(errStr, "not found") || strings.Contains(errStr, "empty") {
			os.RemoveAll(tempDir)
			return text, nil, nil
		}
		// For other errors, we still check if any images were written, otherwise we log and return error
	}

	// 4. Read the extracted image file paths
	files, readErr := os.ReadDir(tempDir)
	if readErr != nil {
		os.RemoveAll(tempDir)
		return text, nil, fmt.Errorf("failed to read temp dir: %w", readErr)
	}

	var imagePaths []string
	for _, file := range files {
		if !file.IsDir() {
			imagePaths = append(imagePaths, filepath.Join(tempDir, file.Name()))
		}
	}

	// If no images were extracted, delete the tempDir
	if len(imagePaths) == 0 {
		os.RemoveAll(tempDir)
	}

	return text, imagePaths, nil
}
