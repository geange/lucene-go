package analysis

import "github.com/geange/lucene-go/core/tokenattr"

// StopFilter Removes stop words from a token stream.
type StopFilter struct {
	*FilteringTokenFilterBase

	stopWords *CharArraySet
	termAtt   tokenattr.CharTermAttribute
}

func (r *StopFilter) Accept() (bool, error) {
	bytes := []byte(string(r.termAtt.Buffer()))

	return !r.stopWords.Contain(bytes), nil
}

func NewStopFilter(in TokenStream, stopWords *CharArraySet) *StopFilter {
	stopFilter := &StopFilter{
		FilteringTokenFilterBase: nil,
		stopWords:                stopWords,
		termAtt:                  in.AttributeSource().CharTerm(),
	}

	impl := NewFilteringTokenFilterImp(stopFilter, in)
	stopFilter.FilteringTokenFilterBase = impl

	return stopFilter
}
