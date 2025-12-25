// internal/parser/postings.go
package parser

import (
	"sort"
	"strings"
)

// OrderPostingsBySign sorts postings so negative amounts appear before positives.
func OrderPostingsBySign(postings []Posting) {
	if len(postings) < 2 {
		return
	}
	sort.SliceStable(postings, func(i, j int) bool {
		iNeg := strings.HasPrefix(postings[i].Amount.Value, "-")
		jNeg := strings.HasPrefix(postings[j].Amount.Value, "-")
		return iNeg && !jNeg
	})
}
