package search

import (
	"github.com/geange/lucene-go/core/types"
)

// DisjunctionScorer
// Base class for Scorers that score disjunctions.
type DisjunctionScorer struct {
	*BaseScorer

	needsScores bool

	subScorers *DisiPriorityQueue

	approximation types.DocIdSetIterator

	twoPhase *TwoPhase
}

var _ TwoPhaseIterator = &TwoPhase{}

type TwoPhase struct {
}

func (t *TwoPhase) Approximation() types.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}

func (t *TwoPhase) Matches() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TwoPhase) MatchCost() float64 {
	//TODO implement me
	panic("implement me")
}
