// internal/shared/validate.go
package shared

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/jason-riddle/ledger-go/internal/parser"
)

// ValidateTransactions checks that transactions balance per currency.
func ValidateTransactions(txs []*parser.Transaction) error {
	slog.Debug("Starting transaction validation", "transaction_count", len(txs))
	for i, tx := range txs {
		balances := make(map[string]float64)
		for _, posting := range tx.Postings {
			currency := posting.Amount.Currency
			amount, err := strconv.ParseFloat(posting.Amount.Value, 64)
			if err != nil {
				slog.Error("Invalid amount in transaction", "tx_index", i, "amount", posting.Amount.Value, "error", err)
				return fmt.Errorf("invalid amount %s: %w", posting.Amount.Value, err)
			}
			balances[currency] += amount
		}
		for currency, balance := range balances {
			if balance != 0 {
				slog.Error("Transaction does not balance", "tx_index", i, "currency", currency, "balance", balance)
				return fmt.Errorf("transaction does not balance for currency %s: %.2f", currency, balance)
			}
		}
	}
	slog.Debug("All transactions validated successfully")
	return nil
}
