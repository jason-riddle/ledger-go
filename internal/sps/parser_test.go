// internal/sps/parser_test.go
package sps_test

import (
	"testing"

	"github.com/jason-riddle/ledger-go/internal/sps"
)

func TestParser_Parse(t *testing.T) {
	// Sample SPS text
	text := `01/15 Mortgage Payment -500.00
02/15 Mortgage Payment -500.00`

	parser := sps.NewParser()
	txs, err := parser.Parse(text)
	if err != nil {
		t.Fatal(err)
	}

	// Check number of transactions
	if len(txs) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(txs))
	}

	// Check first transaction
	tx := txs[0]
	if tx.Date != "2024-01-15" || tx.Payee != "SPS Mortgage" {
		t.Errorf("Unexpected transaction: %+v", tx)
	}

	if len(tx.Postings) != 2 {
		t.Errorf("Expected 2 postings, got %d", len(tx.Postings))
	}
}
