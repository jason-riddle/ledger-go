// internal/validation/library.go
package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/jason-riddle/ledger-go/internal/parser"
	"github.com/robinvdvleuten/beancount/loader"
)

// ValidateWithLibrary validates transactions using robin-beancount library.
// TODO: Debug import issues and integrate properly.
func ValidateWithLibrary(txs []*parser.Transaction) error {
	// Collect all accounts and currencies
	accountSet := make(map[string]bool)
	currencySet := make(map[string]bool)
	for _, tx := range txs {
		for _, p := range tx.Postings {
			accountSet[p.Account] = true
			currencySet[p.Amount.Currency] = true
		}
	}

	// Add open directives
	var lines []string
	for account := range accountSet {
		for currency := range currencySet {
			line := fmt.Sprintf("2024-01-01 open %s %s", account, currency)
			lines = append(lines, line)
		}
	}
	lines = append(lines, "")

	// Add transactions
	for _, tx := range txs {
		line := fmt.Sprintf("%s * \"%s\"", tx.Date, tx.Payee)
		lines = append(lines, line)
		for _, p := range tx.Postings {
			line := fmt.Sprintf("  %s  %s %s", p.Account, p.Amount.Value, p.Amount.Currency)
			lines = append(lines, line)
		}
		lines = append(lines, "")
	}
	text := strings.Join(lines, "\n")

	// Validate using robin-beancount
	ldr := loader.New()
	_, err := ldr.LoadBytes(context.Background(), "validation.bean", []byte(text))
	return err
}
