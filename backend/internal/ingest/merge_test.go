package ingest

import (
	"testing"
)

func TestMergePageContent(t *testing.T) {
	t.Run("digital page with no images", func(t *testing.T) {
		res := MergePageContent(1, "digital", "Hello World", nil)
		if res.PageNum != 1 {
			t.Errorf("expected page 1, got %d", res.PageNum)
		}
		if res.SourceType != "digital" {
			t.Errorf("expected digital source type, got %s", res.SourceType)
		}
		if res.HasImages {
			t.Error("expected hasImages to be false")
		}
		if res.MergedText != "Hello World" {
			t.Errorf("expected merged text to be 'Hello World', got %q", res.MergedText)
		}
	})

	t.Run("scanned page with single image", func(t *testing.T) {
		res := MergePageContent(2, "scanned", "OCR text", []string{"A description of a chart"})
		if res.PageNum != 2 {
			t.Errorf("expected page 2, got %d", res.PageNum)
		}
		if res.SourceType != "scanned" {
			t.Errorf("expected scanned source type, got %s", res.SourceType)
		}
		if !res.HasImages {
			t.Error("expected hasImages to be true")
		}
		expectedText := "OCR text\n\n--- Page Image Description ---\nA description of a chart"
		if res.MergedText != expectedText {
			t.Errorf("expected merged text to match, got %q", res.MergedText)
		}
	})

	t.Run("digital page with multiple images", func(t *testing.T) {
		res := MergePageContent(3, "digital", "Base text", []string{"Image description 1", "Image description 2"})
		if !res.HasImages {
			t.Error("expected hasImages to be true")
		}
		expectedText := "Base text\n\n--- Embedded Image Descriptions ---\n[Image 1]: Image description 1\n\n[Image 2]: Image description 2\n\n"
		if res.MergedText != expectedText {
			t.Errorf("expected merged text to match, got %q", res.MergedText)
		}
	})
}
