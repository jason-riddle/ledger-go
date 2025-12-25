// internal/cloverleaf/parser.go
package cloverleaf

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/jason-riddle/ledger-go/internal/parser"
)

type cloverLeafParser struct{}

// NewParser creates a new CloverLeaf parser.
func NewParser() parser.Parser {
	return &cloverLeafParser{}
}

// Parse extracts transactions from CloverLeaf statement text.
func (p *cloverLeafParser) Parse(text string) ([]*parser.Transaction, error) {
	slog.Debug("Starting CloverLeaf parsing", "text_length", len(text))
	var txs []*parser.Transaction

	// Add opening balance transaction
	openingTx := p.createOpeningBalance(text)
	if openingTx != nil {
		txs = append(txs, openingTx)
		slog.Debug("Added opening balance transaction")
	}

	// Regex for transaction lines: desc date increase decrease
	re := regexp.MustCompile(`(.+?)\s+(\d{2}-\d{2}-\d{4})\s+[\$]?([\d,]+\.\d{2}|0\.00)\s+[\$]?([\d,]+\.\d{2}|0\.00)`)
	matches := re.FindAllStringSubmatch(text, -1)
	slog.Debug("Found potential transaction lines", "count", len(matches))

	for _, match := range matches {
		desc := strings.TrimSpace(match[1])
		dateStr := match[2]
		increaseStr := strings.ReplaceAll(match[3], ",", "")
		decreaseStr := strings.ReplaceAll(match[4], ",", "")

		// Convert MM-DD-YYYY to YYYY-MM-DD
		parts := strings.Split(dateStr, "-")
		if len(parts) == 3 {
			dateStr = fmt.Sprintf("%s-%s-%s", parts[2], parts[0], parts[1])
		}

		// Skip if both amounts are 0 or if desc contains certain words
		if increaseStr == "0.00" && decreaseStr == "0.00" {
			continue
		}
		if strings.Contains(desc, "Tenant") || strings.Contains(desc, "Layla") {
			continue
		}

		// Determine payee and accounts based on desc (simplified)
		payee := p.mapPayee(desc)
		account := p.mapAccount(desc)

		var postings []parser.Posting
		if increaseStr != "0.00" {
			// Increase in PM account (income or reduction)
			postings = []parser.Posting{
				{Account: account, Amount: parser.Amount{Value: fmt.Sprintf("-%s", increaseStr), Currency: "USD"}},
				{Account: "Assets:Property-Management:CloverLeaf-PM", Amount: parser.Amount{Value: increaseStr, Currency: "USD"}},
			}
		} else if decreaseStr != "0.00" {
			// Decrease in PM account (expense or distribution)
			postings = []parser.Posting{
				{Account: "Assets:Property-Management:CloverLeaf-PM", Amount: parser.Amount{Value: fmt.Sprintf("-%s", decreaseStr), Currency: "USD"}},
				{Account: account, Amount: parser.Amount{Value: decreaseStr, Currency: "USD"}},
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

	slog.Info("Completed CloverLeaf parsing", "transactions", len(txs))
	return txs, nil
}

// createOpeningBalance creates the opening balance transaction
func (p *cloverLeafParser) createOpeningBalance(text string) *parser.Transaction {
	// Parse beginning balance
	re := regexp.MustCompile(`Beginning Balance\s*\$?([\d,]+\.\d{2})`)
	match := re.FindStringSubmatch(text)
	if len(match) < 2 {
		return nil
	}
	balance := strings.ReplaceAll(match[1], ",", "")

	return &parser.Transaction{
		Date:      "2025-11-01",
		Payee:     "Opening balance",
		Narration: "Opening balance for property management account",
		Tags:      []string{"#beangulp", "#imported"},
		Links:     nil, // Opening balance has no links in golden
		Postings: []parser.Posting{
			{Account: "Assets:Property-Management:CloverLeaf-PM", Amount: parser.Amount{Value: balance, Currency: "USD"}},
			{Account: "Equity:Opening-Balances", Amount: parser.Amount{Value: fmt.Sprintf("-%s", balance), Currency: "USD"}},
		},
	}
}

// mapPayee maps description to payee name
func (p *cloverLeafParser) mapPayee(desc string) string {
	switch {
	case strings.Contains(desc, "Rent"):
		return "Tenant"
	case strings.Contains(desc, "Management Fee"):
		return "CloverLeaf Property Management"
	case strings.Contains(desc, "Owner Distribution"):
		return "John Doe"
	case strings.Contains(desc, "Utilities"):
		return "CloverLeaf Property Management"
	case strings.Contains(desc, "Lock Change"):
		return "Contractor"
	case strings.Contains(desc, "General Repairs"):
		return "Contractor"
	case strings.Contains(desc, "EGM Maintenance"):
		return "Contractor"
	default:
		return desc
	}
}

// mapAccount maps description to account
func (p *cloverLeafParser) mapAccount(desc string) string {
	switch {
	case strings.Contains(desc, "Rent"):
		return "Income:Rent:2943-Butterfly-Palm"
	case strings.Contains(desc, "Management Fee"):
		return "Expenses:Management-Fees:2943-Butterfly-Palm"
	case strings.Contains(desc, "Owner Distribution"):
		return "Equity:Owner-Distributions:Owner-Draw"
	case strings.Contains(desc, "Utilities") && strings.Contains(desc, "Electric"):
		return "Expenses:Utilities:Electric:206-Hoover-Ave"
	case strings.Contains(desc, "Utilities") && strings.Contains(desc, "Water"):
		return "Expenses:Utilities:Water:206-Hoover-Ave"
	case strings.Contains(desc, "Lock Change"):
		return "Expenses:Repairs:2943-Butterfly-Palm"
	case strings.Contains(desc, "General Repairs"):
		return "Expenses:Repairs:2943-Butterfly-Palm"
	case strings.Contains(desc, "EGM Maintenance"):
		return "Expenses:Repairs:2943-Butterfly-Palm"
	default:
		return "Expenses:Other"
	}
}

// mapNarration maps description to narration
func (p *cloverLeafParser) mapNarration(desc string) string {
	switch {
	case strings.Contains(desc, "Rent"):
		return "Memo: Rent - Rent (11-2025)"
	case strings.Contains(desc, "Management Fee") && strings.Contains(desc, "for 11/2025"):
		return "Memo: Management Fee Expense - Management Fee Expense for 11/2025"
	case strings.Contains(desc, "Management Fee") && strings.Contains(desc, "Credit for 10/"):
		return "Memo: Management Fee Expense - Management Fee Expense Credit for 10/"
	case strings.Contains(desc, "Owner Distribution"):
		return "Memo: Owner Distribution"
	case strings.Contains(desc, "Utilities") && strings.Contains(desc, "Electric/Gas Bill - 10/11/25"):
		return "Memo: Utilities - Electric/Gas Bill - 10/11/25 to 11/12/25"
	case strings.Contains(desc, "Utilities") && strings.Contains(desc, "Electric"):
		return "Memo: Utilities - Electric/Gas Bill - 10/6/25 to 10/10/25"
	case strings.Contains(desc, "Utilities") && strings.Contains(desc, "Water"):
		return "Memo: Utilities - Water Bill - 10/6/25 to 10/22/25"
	case strings.Contains(desc, "Lock Change"):
		return "Memo: Lock Change - Lock change - Landlord compliance"
	case strings.Contains(desc, "General Repairs"):
		return "Memo: General Repairs - Trash out"
	case strings.Contains(desc, "EGM Maintenance"):
		return "Memo: EGM Maintenance - Measured current fridge 65in Heigh, 28in width,"
	default:
		return ""
	}
}

// mapLinks returns the standard links
func (p *cloverLeafParser) mapLinks() map[string]string {
	return map[string]string{
		"paperless_bill_invoice_receipt_url": "No doc",
		"property_manager_bill_url":          "No bill",
		"additional_url":                     "No additional url",
		"comments":                           "No comments",
		"work_order_url":                     "Not a work order",
	}
}
