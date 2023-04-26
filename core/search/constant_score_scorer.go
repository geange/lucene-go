package search

import "github.com/geange/lucene-go/core/index"

var _ Scorer = &ConstantScoreScorer{}

type ConstantScoreScorer struct {
	*ScorerDefault

	score            float32
	scoreMode        ScoreMode
	approximation    index.DocIdSetIterator
	twoPhaseIterator TwoPhaseIterator
	disi             index.DocIdSetIterator
}

func (c *ConstantScoreScorer) Score() (float32, error) {
	return c.score, nil
}

func (c *ConstantScoreScorer) DocID() int {
	return c.disi.DocID()
}

func (c *ConstantScoreScorer) Iterator() index.DocIdSetIterator {
	return c.disi
}

func (c *ConstantScoreScorer) GetMaxScore(upTo int) (float32, error) {
	return c.score, nil
}

// NewConstantScoreScorer
// Constructor based on a DocIdSetIterator which will be used to drive iteration.
// Two phase iteration will not be supported.
//
// Params:
//
//	weight – the parent weight
//	score – the score to return on each document
//	scoreMode – the score mode
//	disi – the iterator that defines matching documents
func NewConstantScoreScorer(weight Weight, score float32,
	scoreMode *ScoreMode, disi index.DocIdSetIterator) (*ConstantScoreScorer, error) {

	if scoreMode.Equal(TOP_SCORES) {
		//
	}

	scorer := &ConstantScoreScorer{
		score:            score,
		scoreMode:        *scoreMode,
		approximation:    disi,
		twoPhaseIterator: nil,
		disi:             disi,
	}
	scorer.ScorerDefault = NewScorer(weight)

	return scorer, nil
}
