package search

import "github.com/geange/lucene-go/core/index"

var _ Scorer = &BlockMaxConjunctionScorer{}

type BlockMaxConjunctionScorer struct {
	*ScorerDefault

	scorers            []Scorer
	approximations     []index.DocIdSetIterator
	twoPhases          []TwoPhaseIterator
	maxScorePropagator *MaxScoreSumPropagator
	minScore           float64
}

func NewBlockMaxConjunctionScorer(weight Weight, scorersList []Scorer) *BlockMaxConjunctionScorer {
	panic("")
}

func (b *BlockMaxConjunctionScorer) Score() (float64, error) {
	score := 0.0
	for _, scorer := range b.scorers {
		num, err := scorer.Score()
		if err != nil {
			return 0, err
		}
		score += num
	}
	return score, nil
}

func (b *BlockMaxConjunctionScorer) DocID() int {
	return b.scorers[0].DocID()
}

func (b *BlockMaxConjunctionScorer) Iterator() index.DocIdSetIterator {
	if len(b.twoPhases) == 0 {
		return b.approximation()
	}
	return AsDocIdSetIterator(b.twoPhaseIterator())
}

func (b *BlockMaxConjunctionScorer) GetMaxScore(upTo int) (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxConjunctionScorer) twoPhaseIterator() TwoPhaseIterator {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxConjunctionScorer) approximation() index.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}
