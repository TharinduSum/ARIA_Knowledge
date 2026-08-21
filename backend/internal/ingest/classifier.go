package ingest

import (
	"bytes"
	"fmt"

	"github.com/ledongthuc/pdf"
)

const DefaultThreshold = 25

// IsScannedPage classifies a PDF page by extracting its text layer.
// If the character count (runes) is below the threshold, it is treated as a scanned page.
// It returns a boolean indicating if it's scanned, the extracted text, and any error.
func IsScannedPage(page pdf.Page, threshold int) (bool, string, error) {
	if page.V.IsNull() {
		return true, "", nil
	}

	rows, err := page.GetTextByRow()
	if err != nil {
		// If we can't extract rows, treat as scanned
		return true, "", fmt.Errorf("failed to get text by row: %w", err)
	}

	var buf bytes.Buffer
	for _, row := range rows {
		for _, text := range row.Content {
			buf.WriteString(text.S)
			buf.WriteString(" ")
		}
		buf.WriteString("\n")
	}

	extractedText := buf.String()
	charCount := len([]rune(extractedText))

	isScanned := charCount < threshold
	return isScanned, extractedText, nil
}
