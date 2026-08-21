package ingest

import (
	"regexp"
	"strings"
)

var (
	// hyphenRegex removes line breaks with hyphens (e.g., sur-\nface -> surface)
	hyphenRegex = regexp.MustCompile(`(\w+)-\s*\n\s*(\w+)`)
	// spaceRegex collapses extra spaces and newlines into a single space
	spaceRegex = regexp.MustCompile(`\s+`)
)

// cleanExtractedText cleans up hyphens at line breaks and collapses unnecessary multiple spaces/newlines.
func cleanExtractedText(text string) string {
	// 1. Line breaks මැද තියෙන hyphens ඉවත් කිරීම (sur-\nface -> surface)
	text = hyphenRegex.ReplaceAllString(text, "$1$2")

	// 2. අනවශ්ය extra spaces/newlines තනි space එකක් බවට පත් කිරීම
	text = spaceRegex.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}
