// cmd/lgo/main.go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/jason-riddle/ledger-go/internal/cloverleaf"
	"github.com/jason-riddle/ledger-go/internal/shared"
)

var (
	pdfPath   = flag.String("pdf-path", "", "Path to the PDF statement file to process")
	outputDir = flag.String("output-dir", ".", "Directory to write generated .bean files")
	verbose   = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Parse()

	// Setup logging
	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	if *pdfPath == "" {
		fmt.Fprintf(os.Stderr, "Error: --pdf-path is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Extract text from PDF
	text, err := shared.ExtractText(*pdfPath)
	if err != nil {
		slog.Error("Failed to extract text", "error", err)
		os.Exit(1)
	}

	// Parse transactions
	parser := cloverleaf.NewParser()
	txs, err := parser.Parse(text)
	if err != nil {
		slog.Error("Failed to parse transactions", "error", err)
		os.Exit(1)
	}

	// Validate transactions
	if err := shared.ValidateTransactions(txs); err != nil {
		slog.Error("Validation failed", "error", err)
		os.Exit(1)
	}

	// Write output files
	if err := shared.WriteBeanFiles(*outputDir, *pdfPath, txs); err != nil {
		slog.Error("Failed to write files", "error", err)
		os.Exit(1)
	}

	slog.Info("Successfully processed statement", "pdf", *pdfPath)
}
