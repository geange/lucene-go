package analysis

import "github.com/geange/lucene-go/core/tokenattributes"

// StopFilter Removes stop words from a token stream.
type StopFilter struct {
	*FilteringTokenFilterImp

	stopWords *CharArraySet
	termAtt   tokenattributes.CharTermAttribute
}

func (r *StopFilter) Accept() (bool, error) {
	bytes := []byte(string(r.termAtt.Buffer()))

	return !r.stopWords.Contain(bytes), nil
}

func NewStopFilter(in TokenStream, stopWords *CharArraySet) *StopFilter {
	stopFilter := &StopFilter{
		FilteringTokenFilterImp: nil,
		stopWords:               stopWords,
		termAtt:                 in.AttributeSource().CharTerm(),
	}

	stopFilter.FilteringTokenFilterImp =
		NewFilteringTokenFilterImp(stopFilter.accept, in)

	return stopFilter
}
