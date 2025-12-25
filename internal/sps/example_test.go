// internal/sps/example_test.go
package sps_test

import (
	"fmt"

	"github.com/jason-riddle/ledger-go/internal/sps"
)

// ExampleNewParser demonstrates creating and using an SPS parser.
func ExampleNewParser() {
	parser := sps.NewParser()
	text := `01/15 Mortgage Payment -500.00`

	txs, err := parser.Parse(text)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, tx := range txs {
		fmt.Printf("%s * \"%s\" \"%s\" beangulp imported\n", tx.Date, tx.Payee, tx.Narration)
		for _, p := range tx.Postings {
			fmt.Printf("  %s  %s %s\n", p.Account, p.Amount.Value, p.Amount.Currency)
		}
		fmt.Println()
	}

	// Output:
	// 2024-01-15 * "SPS Mortgage" "Mortgage Payment" beangulp imported
	//   Liabilities:Mortgages:SPS  -500.00 USD
	//   Expenses:Mortgage-Interest:SPS  500.00 USD
}
