// internal/shared/write_test.go
package shared

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jason-riddle/ledger-go/internal/parser"
)

func TestWriteBeanFiles(t *testing.T) {
	tempDir := t.TempDir()

	txs := []*parser.Transaction{
		{
			Date:  "2024-01-01",
			Payee: "Test Payee",
			Postings: []parser.Posting{
				{Account: "Assets:Checking", Amount: parser.Amount{Value: "100.00", Currency: "USD"}},
				{Account: "Expenses:Other", Amount: parser.Amount{Value: "-100.00", Currency: "USD"}},
			},
		},
	}

	pdfPath := filepath.Join(tempDir, "test.pdf")
	err := WriteBeanFiles(tempDir, pdfPath, txs)
	if err != nil {
		t.Fatalf("WriteBeanFiles failed: %v", err)
	}

	// Check if files exist
	beanPath := filepath.Join(tempDir, "test.bean")
	if _, err := os.Stat(beanPath); os.IsNotExist(err) {
		t.Errorf("Bean file not created")
	}

	balancesPath := filepath.Join(tempDir, "test.balances.bean")
	if _, err := os.Stat(balancesPath); os.IsNotExist(err) {
		t.Errorf("Balances file not created")
	}

	importPath := filepath.Join(tempDir, "test.import.bean")
	if _, err := os.Stat(importPath); os.IsNotExist(err) {
		t.Errorf("Import file not created")
	}

	// Check content
	content, err := os.ReadFile(beanPath)
	if err != nil {
		t.Fatalf("Failed to read bean file: %v", err)
	}
	accountWidth, amountWidth := ComputePostingWidths(txs)
	expected := "2024-01-01 * \"Test Payee\"\n" +
		formatPostingLine(txs[0].Postings[0], accountWidth, amountWidth) + "\n" +
		formatPostingLine(txs[0].Postings[1], accountWidth, amountWidth) + "\n\n"
	if string(content) != expected {
		t.Errorf("Bean file content mismatch: got %q, want %q", string(content), expected)
	}
}
