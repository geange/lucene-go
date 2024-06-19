package search

import (
	"github.com/geange/lucene-go/core/interface/search"
	"github.com/geange/lucene-go/core/types"
)

// A FilterScorer contains another Scorer, which it uses as its basic source of data,
// possibly transforming the data along the way or providing additional functionality.
// The class FilterScorer itself simply implements all abstract methods of Scorer with versions
// that pass all requests to the contained scorer. Subclasses of FilterScorer may further
// override some of these methods and may also provide additional methods and fields.
type FilterScorer struct {
	*BaseScorer

	in search.Scorer
}

func newFilterScorer(in search.Scorer) *FilterScorer {
	return &FilterScorer{
		BaseScorer: NewScorer(in.GetWeight()),
		in:         in,
	}
}

func (f *FilterScorer) Score() (float64, error) {
	return f.in.Score()
}

func (f *FilterScorer) DocID() int {
	return f.in.DocID()
}

func (f *FilterScorer) Iterator() types.DocIdSetIterator {
	return f.in.Iterator()
}

func (f *FilterScorer) TwoPhaseIterator() search.TwoPhaseIterator {
	return f.in.TwoPhaseIterator()
}
