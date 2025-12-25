// internal/parser/interface.go
package parser

// Transaction represents a Beancount transaction.
type Transaction struct {
	Date      string
	Payee     string
	Narration string
	Tags      []string
	Links     map[string]string
	Postings  []Posting
}

// Posting represents a transaction posting.
type Posting struct {
	Account string
	Amount  Amount
}

// Amount represents a monetary amount.
type Amount struct {
	Value    string
	Currency string
}

// Parser defines the interface for parsing statement text into transactions.
type Parser interface {
	Parse(text string) ([]*Transaction, error)
}
