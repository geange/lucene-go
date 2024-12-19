package analysis

import (
	"strings"

	"github.com/geange/lucene-go/core/util/attribute"
)

// LowerCaseFilter Normalizes token text to lower case.
type LowerCaseFilter struct {
	*BaseTokenFilter

	termAtt attribute.CharTermAttr
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
		lower := strings.ToLower(r.termAtt.GetString())
		_ = r.termAtt.Reset()
		if err := r.termAtt.AppendString(lower); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}
