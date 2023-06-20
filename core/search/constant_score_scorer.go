package search

import "github.com/geange/lucene-go/core/index"

var _ Scorer = &ConstantScoreScorer{}

type ConstantScoreScorer struct {
	*ScorerDefault

	score            float64
	scoreMode        ScoreMode
	approximation    index.DocIdSetIterator
	twoPhaseIterator TwoPhaseIterator
	disi             index.DocIdSetIterator
}

func (c *ConstantScoreScorer) Score() (float64, error) {
	return c.score, nil
}

func (c *ConstantScoreScorer) DocID() int {
	return c.disi.DocID()
}

func (c *ConstantScoreScorer) Iterator() index.DocIdSetIterator {
	return c.disi
}

func (c *ConstantScoreScorer) GetMaxScore(upTo int) (float64, error) {
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
func NewConstantScoreScorer(weight Weight, score float64,
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

func NewConstantScoreScorerV1(weight Weight, score float64,
	scoreMode *ScoreMode, twoPhaseIterator TwoPhaseIterator) (*ConstantScoreScorer, error) {

	scorer := &ConstantScoreScorer{
		score:     score,
		scoreMode: *scoreMode,
	}

	if scoreMode.Equal(TOP_SCORES) {
		scorer.approximation = NewStartDISIWrapper(twoPhaseIterator.Approximation())
		scorer.twoPhaseIterator = &constantTwoPhaseIterator{
			approximation:    scorer.approximation,
			twoPhaseIterator: twoPhaseIterator,
		}
	} else {
		scorer.approximation = twoPhaseIterator.Approximation()
		scorer.twoPhaseIterator = twoPhaseIterator
	}
	scorer.ScorerDefault = NewScorer(weight)
	scorer.disi = AsDocIdSetIterator(twoPhaseIterator)
	return scorer, nil
}

var _ TwoPhaseIterator = &constantTwoPhaseIterator{}

type constantTwoPhaseIterator struct {
	approximation    index.DocIdSetIterator
	twoPhaseIterator TwoPhaseIterator
}

func (t *constantTwoPhaseIterator) Approximation() index.DocIdSetIterator {
	return t.approximation
}

func (t *constantTwoPhaseIterator) Matches() (bool, error) {
	return t.twoPhaseIterator.Matches()
}

func (t *constantTwoPhaseIterator) MatchCost() float64 {
	return t.twoPhaseIterator.MatchCost()
}
