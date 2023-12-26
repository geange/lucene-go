package analysis

// StopWordAnalyzer Base class for Analyzers that need to make use of stopword sets.
// Since: 3.1
type StopWordAnalyzer interface {
	Analyzer
}

type DefStopWordAnalyzer struct {
	stopWords *CharArraySet
}

func NewStopWordAnalyzer(stopWords *CharArraySet) *DefStopWordAnalyzer {
	return &DefStopWordAnalyzer{
		stopWords: stopWords,
	}
}

// GetStopWordSet Returns the analyzer's stopWord set or an empty set if the analyzer has no stopWords
// Returns: the analyzer's stopWord set or an empty set if the analyzer has no stopWords
func (r *DefStopWordAnalyzer) GetStopWordSet() *CharArraySet {
	return r.stopWords
}
