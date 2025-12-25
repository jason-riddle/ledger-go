// internal/shared/write.go
package shared

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jason-riddle/ledger-go/internal/parser"
)

// WriteBeanFiles writes the main .bean file and placeholder balances/import files.
func WriteBeanFiles(outputDir, pdfPath string, txs []*parser.Transaction) error {
	baseName := strings.TrimSuffix(filepath.Base(pdfPath), ".pdf")
	slog.Debug("Writing output files", "base_name", baseName, "output_dir", outputDir)

	// Write main .bean file
	beanPath := filepath.Join(outputDir, baseName+".bean")
	file, err := os.Create(beanPath)
	if err != nil {
		slog.Error("Failed to create bean file", "path", beanPath, "error", err)
		return err
	}
	defer file.Close()

	accountWidth, amountWidth := ComputePostingWidths(txs)
	for _, tx := range txs {
		if tx.Directive == "balance" {
			fmt.Fprintln(file, FormatBalanceLine(tx, accountWidth, amountWidth))
			fmt.Fprintln(file)
			continue
		}
		fmt.Fprintf(file, "%s * \"%s\"", tx.Date, tx.Payee)
		if tx.Narration != "" {
			fmt.Fprintf(file, " \"%s\"", tx.Narration)
		}
		if len(tx.Tags) > 0 {
			fmt.Fprintf(file, " %s", strings.Join(tx.Tags, " "))
		}
		fmt.Fprintln(file)
		// Sort links for consistent output
		var keys []string
		for k := range tx.Links {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(file, "  %s: \"%s\"\n", key, tx.Links[key])
		}
		for _, p := range tx.Postings {
			fmt.Fprintln(file, formatPostingLine(p, accountWidth, amountWidth))
		}
		fmt.Fprintln(file)
	}
	slog.Info("Wrote main bean file", "path", beanPath, "transactions", len(txs))

	// Placeholder for balances and import files (extend later)
	balancesPath := filepath.Join(outputDir, baseName+".balances.bean")
	if err := os.WriteFile(balancesPath, []byte("; Balances placeholder\n"), 0644); err != nil {
		slog.Error("Failed to write balances file", "path", balancesPath, "error", err)
		return err
	}
	slog.Debug("Wrote balances placeholder file", "path", balancesPath)

	importPath := filepath.Join(outputDir, baseName+".import.bean")
	if err := os.WriteFile(importPath, []byte("; Import placeholder\n"), 0644); err != nil {
		slog.Error("Failed to write import file", "path", importPath, "error", err)
		return err
	}
	slog.Debug("Wrote import placeholder file", "path", importPath)

	return nil
}
