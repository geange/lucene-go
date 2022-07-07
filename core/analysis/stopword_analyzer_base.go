package analysis

// StopWordAnalyzerBase Base class for Analyzers that need to make use of stopword sets.
// Since: 3.1
type StopWordAnalyzerBase interface {
	Analyzer
}

type StopWordAnalyzerBaseImp struct {
	stopWords *CharArraySet
}

func NewStopWordAnalyzerBaseImp(stopWords *CharArraySet) *StopWordAnalyzerBaseImp {
	return &StopWordAnalyzerBaseImp{stopWords: stopWords}
}

// GetStopWordSet Returns the analyzer's stopword set or an empty set if the analyzer has no stopwords
// Returns: the analyzer's stopword set or an empty set if the analyzer has no stopwords
func (r *StopWordAnalyzerBaseImp) GetStopWordSet() *CharArraySet {
	return r.stopWords
}
