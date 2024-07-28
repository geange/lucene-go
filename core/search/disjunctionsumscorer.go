package search

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.Scorer = &DisjunctionSumScorer{}

// DisjunctionSumScorer
// A Scorer for OR like queries, counterpart of ConjunctionScorer.
type DisjunctionSumScorer struct {
	*DisjunctionScorer
}

func newDisjunctionScorer(weight index.Weight, subScorers []index.Scorer, scoreMode index.ScoreMode) (*DisjunctionSumScorer, error) {
	panic("")
}

func (d *DisjunctionSumScorer) Score() (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DisjunctionSumScorer) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (d *DisjunctionSumScorer) Iterator() types.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}

func (d *DisjunctionSumScorer) GetMaxScore(upTo int) (float64, error) {
	//TODO implement me
	panic("implement me")
}
