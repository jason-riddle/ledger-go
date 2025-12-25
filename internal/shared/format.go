// internal/shared/format.go
package shared

import (
	"fmt"

	"github.com/jason-riddle/ledger-go/internal/parser"
)

// ComputePostingWidths returns the max account and amount widths across all postings.
func ComputePostingWidths(txs []*parser.Transaction) (int, int) {
	maxAccount := 0
	maxAmount := 0
	for _, tx := range txs {
		for _, p := range tx.Postings {
			if len(p.Account) > maxAccount {
				maxAccount = len(p.Account)
			}
			if len(p.Amount.Value) > maxAmount {
				maxAmount = len(p.Amount.Value)
			}
		}
	}
	return maxAccount, maxAmount
}

// FormatPostingLine aligns postings to a fixed currency column.
func FormatPostingLine(p parser.Posting, accountWidth, amountWidth int) string {
	return formatPostingLine(p, accountWidth, amountWidth)
}

func formatPostingLine(p parser.Posting, accountWidth, amountWidth int) string {
	return fmt.Sprintf("  %-*s  %*s %s", accountWidth, p.Account, amountWidth, p.Amount.Value, p.Amount.Currency)
}
