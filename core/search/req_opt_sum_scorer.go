package search

import "github.com/geange/lucene-go/core/index"

var _ Scorer = &ReqOptSumScorer{}

// ReqOptSumScorer
// A Scorer for queries with a required part and an optional part. Delays skipTo() on the optional part until a score() is needed.
type ReqOptSumScorer struct {
	*ScorerDefault

	reqScorer Scorer
	optScorer Scorer

	reqApproximation index.DocIdSetIterator
	optApproximation index.DocIdSetIterator
	optTwoPhase      TwoPhaseIterator
	approximation    index.DocIdSetIterator
	twoPhase         TwoPhaseIterator

	maxScorePropagator *MaxScoreSumPropagator
	minScore           float64
	reqMaxScore        float64
	optIsRequired      bool
}

func (r *ReqOptSumScorer) Score() (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *ReqOptSumScorer) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (r *ReqOptSumScorer) Iterator() index.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}

func (r *ReqOptSumScorer) GetMaxScore(upTo int) (float64, error) {
	//TODO implement me
	panic("implement me")
}
