package search

import (
	"github.com/geange/lucene-go/core/interface/search"
	"github.com/geange/lucene-go/core/types"
)

var _ search.Scorer = &ConstantScoreScorer{}

type ConstantScoreScorer struct {
	*BaseScorer

	score            float64
	scoreMode        search.ScoreMode
	approximation    types.DocIdSetIterator
	twoPhaseIterator search.TwoPhaseIterator
	disi             types.DocIdSetIterator
}

func (c *ConstantScoreScorer) Score() (float64, error) {
	return c.score, nil
}

func (c *ConstantScoreScorer) DocID() int {
	return c.disi.DocID()
}

func (c *ConstantScoreScorer) Iterator() types.DocIdSetIterator {
	return c.disi
}

func (c *ConstantScoreScorer) GetMaxScore(upTo int) (float64, error) {
	return c.score, nil
}

// NewConstantScoreScorer
// Constructor based on a DocIdSetIterator which will be used to drive iteration.
// Two phase iteration will not be supported.
//
//	weight: the parent weight
//	score: the score to return on each document
//	scoreMode: the score mode
//	disi: the iterator that defines matching documents
func NewConstantScoreScorer(weight search.Weight, score float64,
	scoreMode search.ScoreMode, disi types.DocIdSetIterator) (*ConstantScoreScorer, error) {

	if scoreMode == TOP_SCORES {
		//
	}

	scorer := &ConstantScoreScorer{
		score:            score,
		scoreMode:        scoreMode,
		approximation:    disi,
		twoPhaseIterator: nil,
		disi:             disi,
	}
	scorer.BaseScorer = NewScorer(weight)

	return scorer, nil
}

func NewConstantScoreScorerV1(weight search.Weight, score float64,
	scoreMode search.ScoreMode, twoPhaseIterator search.TwoPhaseIterator) (*ConstantScoreScorer, error) {

	scorer := &ConstantScoreScorer{
		score:     score,
		scoreMode: scoreMode,
	}

	if scoreMode == TOP_SCORES {
		scorer.approximation = NewStartDISIWrapper(twoPhaseIterator.Approximation())
		scorer.twoPhaseIterator = &constantTwoPhaseIterator{
			approximation:    scorer.approximation,
			twoPhaseIterator: twoPhaseIterator,
		}
	} else {
		scorer.approximation = twoPhaseIterator.Approximation()
		scorer.twoPhaseIterator = twoPhaseIterator
	}
	scorer.BaseScorer = NewScorer(weight)
	scorer.disi = AsDocIdSetIterator(twoPhaseIterator)
	return scorer, nil
}

var _ search.TwoPhaseIterator = &constantTwoPhaseIterator{}

type constantTwoPhaseIterator struct {
	approximation    types.DocIdSetIterator
	twoPhaseIterator search.TwoPhaseIterator
}

func (t *constantTwoPhaseIterator) Approximation() types.DocIdSetIterator {
	return t.approximation
}

func (t *constantTwoPhaseIterator) Matches() (bool, error) {
	return t.twoPhaseIterator.Matches()
}

func (t *constantTwoPhaseIterator) MatchCost() float64 {
	return t.twoPhaseIterator.MatchCost()
}
