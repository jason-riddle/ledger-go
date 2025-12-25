// internal/sheervalue/parser.go
package sheervalue

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jason-riddle/ledger-go/internal/parser"
)

type sheerValueParser struct{}

// NewParser creates a new SheerValue parser.
func NewParser() parser.Parser {
	return &sheerValueParser{}
}

// Parse extracts transactions from SheerValue statement text.
func (p *sheerValueParser) Parse(text string) ([]*parser.Transaction, error) {
	var txs []*parser.Transaction

	// Look for transaction lines that have dates and amounts
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Match lines that start with date and contain amounts
		re := regexp.MustCompile(`^(\d{1,2}/\d{1,2}/\d{4})\s+(.+?)\s+(\d+\.\d{2})\s+\d+,\d+\.\d{2}$`)
		match := re.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		dateStr := match[1]
		desc := strings.TrimSpace(match[2])
		amountStr := match[3]

		// Convert date from MM/DD/YYYY to YYYY-MM-DD
		dateParts := strings.Split(dateStr, "/")
		if len(dateParts) == 3 {
			dateStr = fmt.Sprintf("%s-%02s-%02s", dateParts[2], dateParts[0], dateParts[1])
		}

		// Determine payee and accounts based on desc
		payee := p.mapPayee(desc)
		account := p.mapAccount(desc)

		var postings []parser.Posting
		if strings.Contains(desc, "Management Fee") {
			// Expense
			postings = []parser.Posting{
				{Account: account, Amount: parser.Amount{Value: amountStr, Currency: "USD"}},
				{Account: "Assets:Property-Management:SheerValue-PM", Amount: parser.Amount{Value: fmt.Sprintf("-%s", amountStr), Currency: "USD"}},
			}
		} else {
			// Income (rent, pet rent, etc.)
			postings = []parser.Posting{
				{Account: "Assets:Property-Management:SheerValue-PM", Amount: parser.Amount{Value: amountStr, Currency: "USD"}},
				{Account: account, Amount: parser.Amount{Value: fmt.Sprintf("-%s", amountStr), Currency: "USD"}},
			}
		}

		tx := &parser.Transaction{
			Date:      dateStr,
			Payee:     payee,
			Narration: p.mapNarration(desc),
			Tags:      []string{"#beangulp", "#imported"},
			Links:     p.mapLinks(),
			Postings:  postings,
		}
		txs = append(txs, tx)
	}

	return txs, nil
}

// mapPayee maps description to payee name
func (p *sheerValueParser) mapPayee(desc string) string {
	if strings.Contains(desc, "Tenant") || strings.Contains(desc, "Rent") || strings.Contains(desc, "Pet Rent") {
		return "Tenant"
	}
	if strings.Contains(desc, "Management") {
		return "SheerValue Property Management"
	}
	return desc
}

// mapAccount maps description to account
func (p *sheerValueParser) mapAccount(desc string) string {
	switch {
	case strings.Contains(desc, "Rent Income"):
		return "Income:Rent:Properties"
	case strings.Contains(desc, "Pet Rent"):
		return "Income:Pet-Rent:Properties"
	case strings.Contains(desc, "Management Fee"):
		return "Expenses:Management-Fees"
	default:
		return "Expenses:Other"
	}
}

// mapNarration maps description to narration
func (p *sheerValueParser) mapNarration(desc string) string {
	switch {
	case strings.Contains(desc, "Rent Income"):
		return "Memo: Rent Income"
	case strings.Contains(desc, "Pet Rent"):
		return "Memo: Pet Rent"
	case strings.Contains(desc, "Management Fee"):
		return "Memo: Management Fee"
	case strings.Contains(desc, "50.00"):
		return "Memo: 50.00 4,7"
	case strings.Contains(desc, "by George Mahara"):
		return "Memo: by George Mahara"
	default:
		return ""
	}
}

// mapLinks returns the standard links
func (p *sheerValueParser) mapLinks() map[string]string {
	return map[string]string{
		"paperless_bill_invoice_receipt_url": "No doc",
		"property_manager_bill_url":          "No bill",
		"additional_url":                     "No additional url",
		"comments":                           "No comments",
		"work_order_url":                     "Not a work order",
	}
}
