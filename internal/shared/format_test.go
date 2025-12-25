// internal/shared/format_test.go
package shared

import (
	"testing"

	"github.com/jason-riddle/ledger-go/internal/parser"
)

func TestComputePostingWidths(t *testing.T) {
	txs := []*parser.Transaction{
		{
			Postings: []parser.Posting{
				{Account: "Assets:Cash", Amount: parser.Amount{Value: "10.00", Currency: "USD"}},
				{Account: "Expenses:Groceries", Amount: parser.Amount{Value: "-12.50", Currency: "USD"}},
			},
		},
	}

	accountWidth, amountWidth := ComputePostingWidths(txs)
	if accountWidth != minAccountWidth {
		t.Fatalf("account width = %d, want %d", accountWidth, minAccountWidth)
	}
	if amountWidth != len("-12.50") {
		t.Fatalf("amount width = %d, want %d", amountWidth, len("-12.50"))
	}
}

func TestFormatPostingLine(t *testing.T) {
	posting := parser.Posting{
		Account: "Assets",
		Amount:  parser.Amount{Value: "-12.34", Currency: "USD"},
	}

	got := FormatPostingLine(posting, 10, 7)
	want := "  Assets       -12.34 USD"
	if got != want {
		t.Fatalf("FormatPostingLine() = %q, want %q", got, want)
	}
}

func TestFormatBalanceLine(t *testing.T) {
	tests := []struct {
		name string
		tx   *parser.Transaction
		want string
	}{
		{
			name: "default separator",
			tx: &parser.Transaction{
				Date:           "2025-12-24",
				BalanceAccount: "Assets:Cash",
				BalanceAmount:  parser.Amount{Value: "100.00", Currency: "USD"},
			},
			want: "2025-12-24 balance Assets:Cash   100.00 USD",
		},
		{
			name: "cloverleaf separator",
			tx: &parser.Transaction{
				Date:           "2025-12-24",
				BalanceAccount: "Assets:CloverLeaf:Cash",
				BalanceAmount:  parser.Amount{Value: "200.00", Currency: "USD"},
			},
			want: "2025-12-24 balance Assets:CloverLeaf:Cash    200.00 USD",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := FormatBalanceLine(test.tx, 0, 0)
			if got != test.want {
				t.Fatalf("FormatBalanceLine() = %q, want %q", got, test.want)
			}
		})
	}
}
