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

	periodRe := regexp.MustCompile(`Statement Period\s+(\d{2}-\d{2}-\d{4})\s+to\s+(\d{2}-\d{2}-\d{4})`)
	periodLineRe := regexp.MustCompile(`(\d{2}-\d{2}-\d{4})\s+to\s+(\d{2}-\d{2}-\d{4})`)
	beginBalanceRe := regexp.MustCompile(`Beginning Balance\s+(\d{2}-\d{2}-\d{4}).*?\$?\s*([\(]?\d[\d,]*\.\d{2}[)]?)`)
	endBalanceRe := regexp.MustCompile(`Ending Balance.*?\$?\s*([\(]?\d[\d,]*\.\d{2}[)]?)`)

	// Regex for transaction lines: desc date increase decrease
	re := regexp.MustCompile(`(.+?)\s+(\d{2}-\d{2}-\d{4})\s+[\$]?([\d,]+\.\d{2}|0\.00)\s+[\$]?([\d,]+\.\d{2}|0\.00)`)
	lines := strings.Split(text, "\n")
	var currentProperty string
	var matches int
	var statementEndDate string
	inDetails := false
	addedEndingBalance := false
	for _, line := range lines {
		if statementEndDate == "" {
			if periodMatch := periodRe.FindStringSubmatch(line); periodMatch != nil {
				statementEndDate = formatDateDash(periodMatch[2])
			} else if periodLineMatch := periodLineRe.FindStringSubmatch(line); periodLineMatch != nil {
				statementEndDate = formatDateDash(periodLineMatch[2])
			}
		}
		if strings.Contains(line, "TRANSACTION DETAILS") {
			inDetails = true
			continue
		}
		if strings.Contains(line, "OPEN WORK ORDERS") {
			inDetails = false
		}
		if inDetails {
			if beginMatch := beginBalanceRe.FindStringSubmatch(line); beginMatch != nil {
				dateStr := formatDateDash(beginMatch[1])
				amountStr, _ := normalizeAmount(beginMatch[2])
				txs = append(txs, &parser.Transaction{
					Date:           dateStr,
					Directive:      "balance",
					BalanceAccount: "Assets:Property-Management:CloverLeaf-PM",
					BalanceAmount:  parser.Amount{Value: amountStr, Currency: "USD"},
				})
				continue
			}
			if !addedEndingBalance {
				if endMatch := endBalanceRe.FindStringSubmatch(line); endMatch != nil {
					amountStr, _ := normalizeAmount(endMatch[1])
					dateStr := statementEndDate
					if dateStr == "" {
						dateStr = ""
					}
					txs = append(txs, &parser.Transaction{
						Date:           dateStr,
						Directive:      "balance",
						BalanceAccount: "Assets:Property-Management:CloverLeaf-PM",
						BalanceAmount:  parser.Amount{Value: amountStr, Currency: "USD"},
					})
					addedEndingBalance = true
					continue
				}
			}
		}
		if strings.Contains(line, "2943 Butterfly Palm") {
			currentProperty = "2943-Butterfly-Palm"
		} else if strings.Contains(line, "206 Hoover Ave") || strings.Contains(line, "206 Hoover Avenue") {
			currentProperty = "206-Hoover-Ave"
		}

		if !inDetails {
			continue
		}
		match := re.FindStringSubmatch(line)
		if len(match) == 0 {
			continue
		}
		matches++
		desc := strings.TrimSpace(match[1])
		dateStr := match[2]
		increaseStr, _ := normalizeAmount(match[3])
		decreaseStr, _ := normalizeAmount(match[4])

		// Convert MM-DD-YYYY to YYYY-MM-DD
		dateStr = formatDateDash(dateStr)

		// Skip if both amounts are 0 or if desc contains certain words
		if increaseStr == "0.00" && decreaseStr == "0.00" {
			continue
		}

		// Determine payee and accounts based on desc (simplified)
		payee := p.mapPayee(desc)
		account := p.mapAccount(desc, currentProperty)

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

	slog.Debug("Found potential transaction lines", "count", matches)
	slog.Info("Completed CloverLeaf parsing", "transactions", len(txs))
	return txs, nil
}

// mapPayee maps description to payee name
func (p *cloverLeafParser) mapPayee(desc string) string {
	switch {
	case strings.Contains(desc, "Rent"):
		return "Tenant"
	case strings.Contains(desc, "Management Fee"):
		return "CloverLeaf Property Management"
	case strings.Contains(desc, "Owner Distribution"):
		return "Jason Riddle"
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
func (p *cloverLeafParser) mapAccount(desc, property string) string {
	switch {
	case strings.Contains(desc, "Rent"):
		if property == "" {
			property = "2943-Butterfly-Palm"
		}
		return fmt.Sprintf("Income:Rent:%s", property)
	case strings.Contains(desc, "Management Fee"):
		if property == "" {
			property = "2943-Butterfly-Palm"
		}
		return fmt.Sprintf("Expenses:Management-Fees:%s", property)
	case strings.Contains(desc, "Owner Distribution"):
		return "Equity:Owner-Distributions:Owner-Draw"
	case strings.Contains(desc, "Utilities") && strings.Contains(desc, "Electric"):
		if property == "" {
			property = "206-Hoover-Ave"
		}
		return fmt.Sprintf("Expenses:Utilities:Electric:%s", property)
	case strings.Contains(desc, "Utilities") && strings.Contains(desc, "Water"):
		if property == "" {
			property = "206-Hoover-Ave"
		}
		return fmt.Sprintf("Expenses:Utilities:Water:%s", property)
	case strings.Contains(desc, "Lock Change"):
		if property == "" {
			property = "2943-Butterfly-Palm"
		}
		return fmt.Sprintf("Expenses:Repairs:%s", property)
	case strings.Contains(desc, "General Repairs"):
		if property == "" {
			property = "2943-Butterfly-Palm"
		}
		return fmt.Sprintf("Expenses:Repairs:%s", property)
	case strings.Contains(desc, "EGM Maintenance"):
		if property == "" {
			property = "2943-Butterfly-Palm"
		}
		return fmt.Sprintf("Expenses:Repairs:%s", property)
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
		"comments": "",
	}
}

func normalizeAmount(amount string) (string, bool) {
	trimmed := strings.TrimSpace(amount)
	trimmed = strings.TrimPrefix(trimmed, "$")
	isNegative := strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, ")")
	trimmed = strings.TrimPrefix(trimmed, "(")
	trimmed = strings.TrimSuffix(trimmed, ")")
	trimmed = strings.ReplaceAll(trimmed, ",", "")
	trimmed = strings.ReplaceAll(trimmed, " ", "")
	if isNegative {
		return fmt.Sprintf("-%s", trimmed), true
	}
	return trimmed, false
}

func formatDateDash(dateStr string) string {
	parts := strings.Split(strings.TrimSpace(dateStr), "-")
	if len(parts) != 3 {
		return strings.TrimSpace(dateStr)
	}
	return fmt.Sprintf("%s-%s-%s", parts[2], parts[0], parts[1])
}
