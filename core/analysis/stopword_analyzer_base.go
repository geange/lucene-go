package analysis

// StopWordAnalyzerBase Base class for Analyzers that need to make use of stopword sets.
// Since: 3.1
type StopWordAnalyzerBase interface {
	Analyzer
}

type StopWordAnalyzerBaseIMP struct {
	stopWords *CharArraySet
}

func NewStopWordAnalyzerBaseIMP(stopWords *CharArraySet) *StopWordAnalyzerBaseIMP {
	return &StopWordAnalyzerBaseIMP{stopWords: stopWords}
}

// GetStopWordSet Returns the analyzer's stopword set or an empty set if the analyzer has no stopwords
// Returns: the analyzer's stopword set or an empty set if the analyzer has no stopwords
func (r *StopWordAnalyzerBaseIMP) GetStopWordSet() *CharArraySet {
	return r.stopWords
}
