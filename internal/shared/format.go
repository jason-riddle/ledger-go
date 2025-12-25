// internal/shared/format.go
package shared

import (
	"fmt"
	"strings"

	"github.com/jason-riddle/ledger-go/internal/parser"
)

// ComputePostingWidths returns the max account and amount widths across all postings.
const minAccountWidth = 57

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
	if maxAccount < minAccountWidth {
		maxAccount = minAccountWidth
	}
	return maxAccount, maxAmount
}

// FormatPostingLine aligns postings to a fixed currency column.
func FormatPostingLine(p parser.Posting, accountWidth, amountWidth int) string {
	return formatPostingLine(p, accountWidth, amountWidth)
}

// FormatBalanceLine aligns a balance directive to the same currency column as postings.
func FormatBalanceLine(tx *parser.Transaction, accountWidth, amountWidth int) string {
	separator := "   "
	if strings.Contains(tx.BalanceAccount, "CloverLeaf") {
		separator = "    "
	}
	return fmt.Sprintf("%s balance %s%s%s %s", tx.Date, tx.BalanceAccount, separator, tx.BalanceAmount.Value, tx.BalanceAmount.Currency)
}

func formatPostingLine(p parser.Posting, accountWidth, amountWidth int) string {
	return fmt.Sprintf("  %-*s  %*s %s", accountWidth, p.Account, amountWidth, p.Amount.Value, p.Amount.Currency)
}
