// internal/shared/extract_test.go
package shared

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractText(t *testing.T) {
	// Create a dummy PDF file (since pdftotext is external, test with existing file)
	// For now, assume a test PDF exists or mock

	// Since pdftotext is required, skip if not available
	if _, err := os.Stat("/usr/bin/pdftotext"); os.IsNotExist(err) {
		t.Skip("pdftotext not available")
	}

	// Use a known test file or create a dummy
	tempDir := t.TempDir()
	pdfPath := filepath.Join(tempDir, "test.pdf")
	// Create a dummy file (not real PDF, but test the command)
	os.WriteFile(pdfPath, []byte("dummy"), 0644)

	_, err := ExtractText(pdfPath)
	// Expect error since not real PDF, but command runs
	if err == nil {
		t.Errorf("Expected error for dummy PDF")
	}
}
