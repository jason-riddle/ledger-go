// internal/shared/extract.go
package shared

import (
	"log/slog"
	"os/exec"
	"strings"
)

// ExtractText runs pdftotext on the PDF and returns the extracted text.
func ExtractText(pdfPath string) (string, error) {
	slog.Debug("Extracting text from PDF", "path", pdfPath)
	cmd := exec.Command("pdftotext", "-layout", pdfPath, "-")
	output, err := cmd.Output()
	if err != nil {
		slog.Error("Failed to run pdftotext", "error", err)
		return "", err
	}
	text := strings.TrimSpace(string(output))
	slog.Debug("Extracted text", "length", len(text))
	return text, nil
}
