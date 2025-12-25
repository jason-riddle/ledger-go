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

	lines := strings.Split(text, "\n")
	beginBalanceRe := regexp.MustCompile(`Beginning cash balance as of\s+(\d{1,2}\s*/\s*\d{1,2}\s*/\s*\d{4}).*?\$?\s*([\(]?\d[\d,\s]*\.\d{2}[)]?)`)
	endBalanceRe := regexp.MustCompile(`Ending cash balance as of\s+(\d{1,2}\s*/\s*\d{1,2}\s*/\s*\d{4}).*?\$?\s*([\(]?\d[\d,\s]*\.\d{2}[)]?)`)
	lineRe := regexp.MustCompile(`^\s*(\d{1,2}/\d{1,2}/\d{4})\s+(2943\s+Butterfly\s+Palm|206\s+Hoover\s+Avenue)\s+(\S+)\s+(Rent\s+Income|Pet\s+Rent|Late\s+Fee|Management|Owner\s+Draw|Repairs)\s+(.+?)\s+([\(]?\d[\d,]*\.\d{2}[)]?)\s+([\(]?\d[\d,]*\.\d{2}[)]?)\s*$`)

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if beginMatch := beginBalanceRe.FindStringSubmatch(line); beginMatch != nil {
			dateStr := formatDateSlash(beginMatch[1])
			amountStr, _ := normalizeAmount(beginMatch[2])
			txs = append(txs, &parser.Transaction{
				Date:           dateStr,
				Directive:      "balance",
				BalanceAccount: "Assets:Property-Management:SheerValue-PM",
				BalanceAmount:  parser.Amount{Value: amountStr, Currency: "USD"},
			})
			continue
		}
		if endMatch := endBalanceRe.FindStringSubmatch(line); endMatch != nil {
			dateStr := formatDateSlash(endMatch[1])
			amountStr, _ := normalizeAmount(endMatch[2])
			txs = append(txs, &parser.Transaction{
				Date:           dateStr,
				Directive:      "balance",
				BalanceAccount: "Assets:Property-Management:SheerValue-PM",
				BalanceAmount:  parser.Amount{Value: amountStr, Currency: "USD"},
			})
			continue
		}

		match := lineRe.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		dateStr := match[1]
		property := cleanSpaces(match[2])
		accountType := cleanSpaces(match[4])
		nameMemo := strings.TrimSpace(match[5])
		amountAbs, isNegative := normalizeAmount(match[6])
		sign := 1
		if isNegative {
			sign = -1
		}
		if accountType == "Management" {
			accountType = "Management Fees"
		}
		propertySlug := propertySlug(property)
		account, accountKind := p.mapAccount(accountType, propertySlug)

		payee := payeeFromMemo(nameMemo, accountKind)
		if accountType == "Management Fees" && i+1 < len(lines) {
			nextLine := lines[i+1]
			if strings.Contains(nextLine, "Fees") {
				payeeSuffix := extractManagementSuffix(nextLine)
				if payeeSuffix != "" {
					payee = cleanSpaces(payee + " " + payeeSuffix)
				}
			}
		}
		reversed := isNegative || strings.Contains(strings.ToUpper(nameMemo), "REVERSED")
		if reversed {
			if suffix := findBySuffix(lines, i+1); suffix != "" && !strings.Contains(payee, suffix) {
				payee = cleanSpaces(payee + " " + suffix)
			}
		}

		narration := fmt.Sprintf("Memo: %s - %s", property, accountType)
		tags := []string{"#beangulp", "#imported"}
		if reversed {
			narration += " - REVERSED"
			tags = append(tags, "#reversed")
		}

		var postings []parser.Posting
		if accountKind == accountKindIncome {
			postings = []parser.Posting{
				{Account: "Assets:Property-Management:SheerValue-PM", Amount: parser.Amount{Value: signedAmount(amountAbs, sign), Currency: "USD"}},
				{Account: account, Amount: parser.Amount{Value: signedAmount(amountAbs, -sign), Currency: "USD"}},
			}
		} else {
			postings = []parser.Posting{
				{Account: account, Amount: parser.Amount{Value: signedAmount(amountAbs, sign), Currency: "USD"}},
				{Account: "Assets:Property-Management:SheerValue-PM", Amount: parser.Amount{Value: signedAmount(amountAbs, -sign), Currency: "USD"}},
			}
		}

		tx := &parser.Transaction{
			Date:      formatDateSlash(dateStr),
			Payee:     payee,
			Narration: narration,
			Tags:      tags,
			Links:     p.mapLinks(),
			Postings:  postings,
		}
		txs = append(txs, tx)
	}

	return txs, nil
}

// mapLinks returns the standard links
func (p *sheerValueParser) mapLinks() map[string]string {
	return map[string]string{
		"comments": "",
	}
}

type accountKind int

const (
	accountKindIncome accountKind = iota
	accountKindExpense
)

func (p *sheerValueParser) mapAccount(accountType, propertySlug string) (string, accountKind) {
	switch accountType {
	case "Rent Income":
		return fmt.Sprintf("Income:Rent:%s", propertySlug), accountKindIncome
	case "Pet Rent":
		return fmt.Sprintf("Income:Pet-Fee:%s", propertySlug), accountKindIncome
	case "Late Fee":
		return fmt.Sprintf("Income:Late-Rent-Fee:%s", propertySlug), accountKindIncome
	case "Management Fees":
		return fmt.Sprintf("Expenses:Management-Fees:%s", propertySlug), accountKindExpense
	case "Owner Draw":
		return "Equity:Owner-Distributions:Owner-Draw", accountKindExpense
	case "Repairs":
		return fmt.Sprintf("Expenses:Repairs:%s", propertySlug), accountKindExpense
	default:
		return "Expenses:Other", accountKindExpense
	}
}

var multiSpaceRe = regexp.MustCompile(`\s{2,}`)

func payeeFromMemo(nameMemo string, accountKind accountKind) string {
	if accountKind == accountKindIncome {
		cleaned := cleanSpaces(nameMemo)
		if idx := strings.Index(strings.ToUpper(cleaned), "REVERSED"); idx != -1 {
			cleaned = cleanSpaces(cleaned[:idx])
		}
		return cleaned
	}
	parts := multiSpaceRe.Split(strings.TrimSpace(nameMemo), -1)
	if len(parts) == 0 {
		return cleanSpaces(nameMemo)
	}
	return cleanSpaces(parts[0])
}

func extractManagementSuffix(line string) string {
	idx := strings.Index(line, "Fees")
	if idx == -1 {
		return ""
	}
	suffix := strings.TrimSpace(line[idx+len("Fees"):])
	return cleanSpaces(suffix)
}

func findBySuffix(lines []string, start int) string {
	for i := start; i < len(lines) && i < start+3; i++ {
		lower := strings.ToLower(lines[i])
		if idx := strings.Index(lower, "by "); idx != -1 {
			return cleanSpaces(lines[i][idx:])
		}
	}
	return ""
}

func normalizeAmount(amount string) (string, bool) {
	trimmed := strings.TrimSpace(amount)
	trimmed = strings.TrimPrefix(trimmed, "$")
	isNegative := strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, ")")
	trimmed = strings.TrimPrefix(trimmed, "(")
	trimmed = strings.TrimSuffix(trimmed, ")")
	if strings.HasPrefix(trimmed, "-") {
		isNegative = true
		trimmed = strings.TrimPrefix(trimmed, "-")
	}
	trimmed = strings.ReplaceAll(trimmed, ",", "")
	trimmed = strings.ReplaceAll(trimmed, " ", "")
	return trimmed, isNegative
}

func signedAmount(amount string, sign int) string {
	if sign < 0 {
		return fmt.Sprintf("-%s", amount)
	}
	return amount
}

func formatDateSlash(dateStr string) string {
	clean := strings.ReplaceAll(dateStr, " ", "")
	parts := strings.Split(clean, "/")
	if len(parts) != 3 {
		return strings.TrimSpace(dateStr)
	}
	return fmt.Sprintf("%s-%02s-%02s", parts[2], parts[0], parts[1])
}

func cleanSpaces(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func propertySlug(property string) string {
	parts := strings.Fields(property)
	for i, part := range parts {
		switch part {
		case "Avenue":
			parts[i] = "Ave"
		case "Street":
			parts[i] = "St"
		case "Road":
			parts[i] = "Rd"
		case "Drive":
			parts[i] = "Dr"
		}
	}
	return strings.Join(parts, "-")
}
