package search

import "github.com/geange/lucene-go/core/index"

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

func (c *ConjunctionScorer) Score() (float32, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConjunctionScorer) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (c *ConjunctionScorer) Iterator() index.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}

func (c *ConjunctionScorer) GetMaxScore(upTo int) (float32, error) {
	//TODO implement me
	panic("implement me")
}

func intersectScorers(scorers []Scorer) index.DocIdSetIterator {

	panic("")
}
