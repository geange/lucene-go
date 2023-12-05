package analysis

import (
	"github.com/geange/lucene-go/core/tokenattr"
	"strings"
)

// LowerCaseFilter Normalizes token text to lower case.
type LowerCaseFilter struct {
	*DefTokenFilter

	termAtt tokenattr.CharTermAttribute
}

func NewLowerCaseFilter(in TokenStream) *LowerCaseFilter {
	filter := LowerCaseFilter{
		DefTokenFilter: NewTokenFilter(in),
		termAtt:        in.AttributeSource().CharTerm(),
	}
	return &filter
}

func (r *LowerCaseFilter) IncrementToken() (bool, error) {
	if ok, err := r.input.IncrementToken(); err != nil {
		return false, err
	} else if ok {
		lower := strings.ToLower(string(r.termAtt.Buffer()))
		r.termAtt.SetEmpty()
		r.termAtt.Append(lower)
		return true, nil
	}
	return false, nil
}
