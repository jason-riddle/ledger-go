// internal/sheervalue/parser_test.go
package sheervalue_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/jason-riddle/ledger-go/internal/shared"
	"github.com/jason-riddle/ledger-go/internal/sheervalue"
)

func TestGoldenFiles(t *testing.T) {
	// Load fixture
	fixturePath := filepath.Join("..", "..", "tests", "fixtures", "sheervalue", "multi_prop", "sheervalue_2025_multi_property_statement.txt")
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}

	// Parse
	parser := sheervalue.NewParser()
	txs, err := parser.Parse(string(fixture))
	if err != nil {
		t.Fatal(err)
	}

	// Format to .bean
	var lines []string
	accountWidth, amountWidth := shared.ComputePostingWidths(txs)
	for _, tx := range txs {
		line := fmt.Sprintf("%s * \"%s\"", tx.Date, tx.Payee)
		if tx.Narration != "" {
			line += fmt.Sprintf(" \"%s\"", tx.Narration)
		}
		line += " " + strings.Join(tx.Tags, " ")
		lines = append(lines, line)
		// Sort link keys for consistent output
		var keys []string
		for key := range tx.Links {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			lines = append(lines, fmt.Sprintf("  %s: \"%s\"", key, tx.Links[key]))
		}
		for _, p := range tx.Postings {
			lines = append(lines, shared.FormatPostingLine(p, accountWidth, amountWidth))
		}
		lines = append(lines, "")
	}
	output := strings.Join(lines, "\n")

	// Load golden
	goldenPath := filepath.Join("..", "..", "tests", "golden", "sheervalue", "sheervalue_2025_multi_property_statement.bean")
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}

	// Compare
	if strings.TrimSpace(output) != strings.TrimSpace(string(golden)) {
		t.Errorf("Output does not match golden file.\nGot:\n%s\n\nWant:\n%s", output, string(golden))
	}
}
