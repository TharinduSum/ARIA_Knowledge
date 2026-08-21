package ingest

import (
	"bytes"
	"fmt"
)

// MergedPage represents the final text and metadata resulting from page ingestion.
type MergedPage struct {
	PageNum    int
	SourceType string // "digital" | "scanned"
	HasImages  bool
	MergedText string
}

// MergePageContent concatenates base text with captions and populates page metadata.
func MergePageContent(pageNum int, sourceType string, baseText string, captions []string) MergedPage {
	var buf bytes.Buffer
	buf.WriteString(cleanExtractedText(baseText))

	if len(captions) > 0 {
		if sourceType == "scanned" {
			if buf.Len() > 0 {
				buf.WriteString("\n\n--- Page Image Description ---\n")
			} else {
				buf.WriteString("--- Page Image Description ---\n")
			}
		} else {
			if buf.Len() > 0 {
				buf.WriteString("\n\n--- Embedded Image Descriptions ---\n")
			} else {
				buf.WriteString("--- Page Image Description ---\n")
			}
		}
		for i, caption := range captions {
			if len(captions) > 1 {
				buf.WriteString(fmt.Sprintf("[Image %d]: %s\n\n", i+1, caption))
			} else {
				buf.WriteString(caption)
			}
		}
	}

	return MergedPage{
		PageNum:    pageNum,
		SourceType: sourceType,
		HasImages:  len(captions) > 0,
		MergedText: buf.String(),
	}
}
