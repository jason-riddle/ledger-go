// internal/sheervalue/example_test.go
package sheervalue_test

import (
	"fmt"
	"strings"

	"github.com/jason-riddle/ledger-go/internal/sheervalue"
)

// ExampleNewParser demonstrates creating and using a SheerValue parser.
func ExampleNewParser() {
	parser := sheervalue.NewParser()
	text := `SheerValue Property Management
Rental Owner Statement
Jason Riddle

Rent Income $1000.00
Management Fee -$50.00`

	txs, err := parser.Parse(text)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, tx := range txs {
		fmt.Printf("%s * \"%s\" %s\n", tx.Date, tx.Payee, strings.Join(tx.Tags, " "))
		for _, p := range tx.Postings {
			fmt.Printf("  %s  %s %s\n", p.Account, p.Amount.Value, p.Amount.Currency)
		}
		fmt.Println()
	}
}
