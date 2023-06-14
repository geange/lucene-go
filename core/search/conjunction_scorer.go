package search

import (
	"github.com/geange/lucene-go/core/index"
	"math"
)

var _ Scorer = &ConjunctionScorer{}

// ConjunctionScorer
// Create a new ConjunctionScorer, note that scorers must be a subset of required.
type ConjunctionScorer struct {
	*ScorerDefault

	disi     index.DocIdSetIterator
	scorers  []Scorer
	required []Scorer
}

func NewConjunctionScorer(weight Weight, scorers []Scorer, required []Scorer) *ConjunctionScorer {
	return &ConjunctionScorer{
		ScorerDefault: NewScorer(weight),
		disi:          intersectScorers(scorers),
		scorers:       scorers,
		required:      required,
	}
}

func (c *ConjunctionScorer) Score() (float64, error) {
	sum := 0.0
	for _, scorer := range c.scorers {
		v, err := scorer.Score()
		if err != nil {
			return 0, err
		}
		sum += v
	}
	return sum, nil
}

func (c *ConjunctionScorer) DocID() int {
	return c.disi.DocID()
}

func (c *ConjunctionScorer) TwoPhaseIterator() TwoPhaseIterator {
	return UnwrapIterator(c.disi)
}

func (c *ConjunctionScorer) Iterator() index.DocIdSetIterator {
	return c.disi
}

func (c *ConjunctionScorer) GetMaxScore(upTo int) (float64, error) {
	switch len(c.scorers) {
	case 0:
		return 0, nil
	case 1:
		return c.scorers[0].GetMaxScore(upTo)
	default:
		return math.Inf(-1), nil
	}
}

func intersectScorers(scorers []Scorer) index.DocIdSetIterator {

	panic("")
}
