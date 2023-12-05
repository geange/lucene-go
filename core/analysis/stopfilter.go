package analysis

import "github.com/geange/lucene-go/core/tokenattr"

// StopFilter Removes stop words from a token stream.
type StopFilter struct {
	*DefFilteringTokenFilter

	stopWords *CharArraySet
	termAtt   tokenattr.CharTermAttribute
}

func (r *StopFilter) Accept() (bool, error) {
	bytes := []byte(string(r.termAtt.Buffer()))

	return !r.stopWords.Contain(bytes), nil
}

func NewStopFilter(in TokenStream, stopWords *CharArraySet) *StopFilter {
	stopFilter := &StopFilter{
		DefFilteringTokenFilter: nil,
		stopWords:               stopWords,
		termAtt:                 in.AttributeSource().CharTerm(),
	}

	impl := NewFilteringTokenFilter(stopFilter, in)
	stopFilter.DefFilteringTokenFilter = impl

	return stopFilter
}
