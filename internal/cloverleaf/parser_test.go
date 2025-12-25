// internal/cloverleaf/parser_test.go
package cloverleaf_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jason-riddle/ledger-go/internal/cloverleaf"
)

func TestGoldenFiles(t *testing.T) {
	// Load fixture
	fixturePath := filepath.Join("..", "..", "tests", "fixtures", "cloverleaf", "cloverleaf_2025-12-11_statement.txt")
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}

	// Parse
	parser := cloverleaf.NewParser()
	txs, err := parser.Parse(string(fixture))
	if err != nil {
		t.Fatal(err)
	}

	// Format to .bean
	var lines []string
	for _, tx := range txs {
		line := fmt.Sprintf("%s * \"%s\"", tx.Date, tx.Payee)
		if tx.Narration != "" {
			line += fmt.Sprintf(" \"%s\"", tx.Narration)
		}
		if len(tx.Tags) > 0 {
			line += " " + strings.Join(tx.Tags, " ")
		}
		lines = append(lines, line)
		// Links in golden order
		if len(tx.Links) > 0 {
			lines = append(lines, "  paperless_bill_invoice_receipt_url: \"No doc\"")
			lines = append(lines, "  property_manager_bill_url: \"No bill\"")
			lines = append(lines, "  additional_url: \"No additional url\"")
			lines = append(lines, "  comments: \"No comments\"")
			lines = append(lines, "  work_order_url: \"Not a work order\"")
		}
		for _, p := range tx.Postings {
			line := fmt.Sprintf("  %s  %s %s", p.Account, p.Amount.Value, p.Amount.Currency)
			lines = append(lines, line)
		}
		lines = append(lines, "")
	}
	output := strings.Join(lines, "\n")

	// Load golden
	goldenPath := filepath.Join("..", "..", "tests", "golden", "cloverleaf_2025-12-11_statement.bean")
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}

	// Compare
	if strings.TrimSpace(output) != strings.TrimSpace(string(golden)) {
		t.Errorf("Output does not match golden file.\nGot:\n%s\n\nWant:\n%s", output, string(golden))
	}
}
