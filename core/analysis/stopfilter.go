package analysis

import "github.com/geange/lucene-go/core/tokenattr"

// StopFilter Removes stop words from a token stream.
type StopFilter struct {
	*BaseFilteringTokenFilter

	stopWords *CharArraySet
	termAtt   tokenattr.CharTermAttr
}

func (r *StopFilter) Accept() (bool, error) {
	bytes := []byte(r.termAtt.GetString())

	return !r.stopWords.Contain(bytes), nil
}

func NewStopFilter(in TokenStream, stopWords *CharArraySet) *StopFilter {
	stopFilter := &StopFilter{
		stopWords: stopWords,
		termAtt:   in.AttributeSource().CharTerm(),
	}
	stopFilter.BaseFilteringTokenFilter = NewFilteringTokenFilter(stopFilter, in)

	return stopFilter
}
