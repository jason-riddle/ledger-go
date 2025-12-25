// internal/sps/parser.go
package sps

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jason-riddle/ledger-go/internal/parser"
)

type spsParser struct{}

// NewParser creates a new SPS parser.
func NewParser() parser.Parser {
	return &spsParser{}
}

// Parse extracts transactions from SPS statement text.
func (p *spsParser) Parse(text string) ([]*parser.Transaction, error) {
	var txs []*parser.Transaction

	// Simple regex for SPS transaction lines (adapt based on actual format)
	re := regexp.MustCompile(`(\d{2}/\d{2})\s+(.+?)\s+(-?\d+\.\d{2})`)
	matches := re.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		dateStr := match[1]
		desc := strings.TrimSpace(match[2])
		amountStr := match[3]

		// Convert MM/DD to YYYY-MM-DD (assume current year)
		dateStr = "2024-" + dateStr[:2] + "-" + dateStr[3:]

		payee := p.mapPayee(desc)
		account := p.mapAccount(desc)

		var postings []parser.Posting
		if strings.HasPrefix(amountStr, "-") {
			// Expense
			amount := strings.TrimPrefix(amountStr, "-")
			postings = []parser.Posting{
				{Account: account, Amount: parser.Amount{Value: amount, Currency: "USD"}},
				{Account: "Liabilities:Mortgages:SPS", Amount: parser.Amount{Value: fmt.Sprintf("-%s", amount), Currency: "USD"}},
			}
		} else {
			// Income or other
			postings = []parser.Posting{
				{Account: "Liabilities:Mortgages:SPS", Amount: parser.Amount{Value: fmt.Sprintf("-%s", amountStr), Currency: "USD"}},
				{Account: account, Amount: parser.Amount{Value: amountStr, Currency: "USD"}},
			}
		}
		parser.OrderPostingsBySign(postings)

		tx := &parser.Transaction{
			Date:      dateStr,
			Payee:     payee,
			Narration: desc,
			Tags:      []string{"beangulp", "imported"},
			Postings:  postings,
		}
		txs = append(txs, tx)
	}

	return txs, nil
}

// mapPayee maps description to payee name
func (p *spsParser) mapPayee(desc string) string {
	switch {
	case strings.Contains(desc, "Mortgage Payment"):
		return "SPS Mortgage"
	default:
		return desc
	}
}

// mapAccount maps description to account
func (p *spsParser) mapAccount(desc string) string {
	switch {
	case strings.Contains(desc, "Mortgage Payment"):
		return "Expenses:Mortgage-Interest:SPS"
	default:
		return "Expenses:Other"
	}
}
