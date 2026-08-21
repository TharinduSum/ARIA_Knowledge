package ingest

import "testing"

func TestCleanExtractedText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "hyphenated word splitting across line break",
			input:    "This is a sur-\nface example.",
			expected: "This is a surface example.",
		},
		{
			name:     "hyphenated word splitting across line break with extra spacing",
			input:    "Multi-\n   page document.",
			expected: "Multipage document.",
		},
		{
			name:     "extra spaces and newlines collapsed",
			input:    "Too   many    spaces   and\n\nnewlines\t  here.",
			expected: "Too many spaces and newlines here.",
		},
		{
			name:     "normal hyphenated words preserved",
			input:    "This is a first-class citizen.",
			expected: "This is a first-class citizen.",
		},
		{
			name:     "trim leading and trailing spaces",
			input:    "   Hello World   \n",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := cleanExtractedText(tt.input)
			if actual != tt.expected {
				t.Errorf("cleanExtractedText(%q) = %q; expected %q", tt.input, actual, tt.expected)
			}
		})
	}
}
