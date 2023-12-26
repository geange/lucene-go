package analysis

import (
	"github.com/geange/lucene-go/core/tokenattr"
	"strings"
)

// LowerCaseFilter Normalizes token text to lower case.
type LowerCaseFilter struct {
	*BaseTokenFilter

	termAtt tokenattr.CharTermAttr
}

func NewLowerCaseFilter(in TokenStream) *LowerCaseFilter {
	filter := LowerCaseFilter{
		BaseTokenFilter: NewBaseTokenFilter(in),
		termAtt:         in.AttributeSource().CharTerm(),
	}
	return &filter
}

func (r *LowerCaseFilter) IncrementToken() (bool, error) {
	ok, err := r.input.IncrementToken()
	if err != nil {
		return false, err
	}

	if ok {
		lower := strings.ToLower(string(r.termAtt.GetString()))
		_ = r.termAtt.Reset()
		r.termAtt.AppendString(lower)
		return true, nil
	}
	return false, nil
}
