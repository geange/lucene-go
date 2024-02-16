package analysis

// StopWordAnalyzer Base class for Analyzers that need to make use of stopword sets.
// Since: 3.1
type StopWordAnalyzer interface {
	Analyzer
}

type BaseStopWordAnalyzer struct {
	stopWords *CharArraySet
}

func NewStopWordAnalyzer(stopWords *CharArraySet) *BaseStopWordAnalyzer {
	return &BaseStopWordAnalyzer{
		stopWords: stopWords,
	}
}

// GetStopWordSet Returns the analyzer's stopWord set or an empty set if the analyzer has no stopWords
// Returns: the analyzer's stopWord set or an empty set if the analyzer has no stopWords
func (r *BaseStopWordAnalyzer) GetStopWordSet() *CharArraySet {
	return r.stopWords
}
