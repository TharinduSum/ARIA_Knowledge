package ingest

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// ExtractScannedPage rasterizes a specific PDF page to an image and runs tesseract OCR.
// It returns the OCR'd text, the path to the rasterized image file, and any error.
// The caller is responsible for deleting the rasterized image file and its parent temp directory when done.
func ExtractScannedPage(pdfPath string, pageNum int) (string, string, error) {
	// 1. Create a temp directory for our files
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("aria-scanned-page-%d-*", pageNum))
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// 2. Rasterize the page using pdftoppm (part of poppler-utils)
	rasterPrefix := filepath.Join(tempDir, "page")
	rasterImgPath := rasterPrefix + ".png"

	pdftoppmCmd := exec.Command("pdftoppm",
		"-png",
		"-r", "150",
		"-f", strconv.Itoa(pageNum),
		"-l", strconv.Itoa(pageNum),
		"-singlefile",
		pdfPath,
		rasterPrefix,
	)

	var pdftoppmStderr bytes.Buffer
	pdftoppmCmd.Stderr = &pdftoppmStderr
	if err := pdftoppmCmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", "", fmt.Errorf("pdftoppm failed (stderr: %q): %w", pdftoppmStderr.String(), err)
	}

	// 3. Run OCR using tesseract CLI
	ocrBase := filepath.Join(tempDir, "ocr_out")
	ocrTxtPath := ocrBase + ".txt"

	tesseractCmd := exec.Command("tesseract",
		rasterImgPath,
		ocrBase,
		"-l", "eng", // Defaulting to English. Make sure tesseract-ocr-eng is installed.
	)

	var tessStderr bytes.Buffer
	tesseractCmd.Stderr = &tessStderr
	if err := tesseractCmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", "", fmt.Errorf("tesseract failed (stderr: %q): %w", tessStderr.String(), err)
	}

	// 4. Read the OCR'd text
	ocrBytes, err := os.ReadFile(ocrTxtPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", "", fmt.Errorf("failed to read OCR output file: %w", err)
	}

	// Remove temporary ocr_out.txt file as it's no longer needed.
	// We keep the rasterImgPath since the caller might need to caption it.
	_ = os.Remove(ocrTxtPath)

	return string(ocrBytes), rasterImgPath, nil
}
