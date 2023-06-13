package search

import "github.com/geange/lucene-go/core/index"

// A FilterScorer contains another Scorer, which it uses as its basic source of data,
// possibly transforming the data along the way or providing additional functionality.
// The class FilterScorer itself simply implements all abstract methods of Scorer with versions
// that pass all requests to the contained scorer. Subclasses of FilterScorer may further
// override some of these methods and may also provide additional methods and fields.
type FilterScorer struct {
	*ScorerDefault

	in Scorer
}

func (f *FilterScorer) Score() (float32, error) {
	return f.in.Score()
}

func (f *FilterScorer) DocID() int {
	return f.in.DocID()
}

func (f *FilterScorer) Iterator() index.DocIdSetIterator {
	return f.in.Iterator()
}

func (f *FilterScorer) TwoPhaseIterator() TwoPhaseIterator {
	return f.in.TwoPhaseIterator()
}
