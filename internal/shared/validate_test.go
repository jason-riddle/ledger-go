// internal/shared/validate_test.go
package shared

import (
	"testing"

	"github.com/jason-riddle/ledger-go/internal/parser"
)

func TestValidateTransactions(t *testing.T) {
	tests := []struct {
		name    string
		txs     []*parser.Transaction
		wantErr bool
	}{
		{
			name: "balanced transactions",
			txs: []*parser.Transaction{
				{
					Postings: []parser.Posting{
						{Amount: parser.Amount{Value: "100.00", Currency: "USD"}},
						{Amount: parser.Amount{Value: "-100.00", Currency: "USD"}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unbalanced transactions",
			txs: []*parser.Transaction{
				{
					Postings: []parser.Posting{
						{Amount: parser.Amount{Value: "100.00", Currency: "USD"}},
						{Amount: parser.Amount{Value: "-50.00", Currency: "USD"}},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransactions(tt.txs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTransactions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
