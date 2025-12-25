// internal/cloverleaf/example_test.go
package cloverleaf_test

import (
	"fmt"

	"github.com/jason-riddle/ledger-go/internal/cloverleaf"
)

// ExampleNewParser demonstrates creating and using a CloverLeaf parser.
func ExampleNewParser() {
	parser := cloverleaf.NewParser()
	text := `CloverLeaf Property Management
Owner Statement
Jason Riddle

Beginning Balance $100.00
Rent Income $500.00`

	txs, err := parser.Parse(text)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, tx := range txs {
		fmt.Printf("%s * \"%s\"\n", tx.Date, tx.Payee)
		for _, p := range tx.Postings {
			fmt.Printf("  %s  %s %s\n", p.Account, p.Amount.Value, p.Amount.Currency)
		}
		fmt.Println()
	}
}
