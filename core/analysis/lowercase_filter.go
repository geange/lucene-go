package analysis

import (
	"github.com/geange/lucene-go/core/tokenattributes"
	"strings"
)

// LowerCaseFilter Normalizes token text to lower case.
type LowerCaseFilter struct {
	*TokenFilterIMP

	termAtt tokenattributes.CharTermAttribute
}

func NewLowerCaseFilter(in TokenStream) *LowerCaseFilter {
	panic("")
}

func (r *LowerCaseFilter) IncrementToken() (bool, error) {
	if ok, err := r.IncrementToken(); err != nil {
		return false, err
	} else if ok {
		lower := strings.ToLower(string(r.termAtt.Buffer()))
		r.termAtt.SetEmpty()
		r.termAtt.Append(lower)
		return true, nil
	}
	return false, nil
}
