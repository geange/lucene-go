package search

import "github.com/geange/lucene-go/core/index"

var _ Scorer = &BlockMaxConjunctionScorer{}

type BlockMaxConjunctionScorer struct {
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
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxConjunctionScorer) SmoothingScore(docId int) (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxConjunctionScorer) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxConjunctionScorer) SetMinCompetitiveScore(minScore float64) error {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxConjunctionScorer) GetChildren() ([]ChildScorable, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxConjunctionScorer) GetWeight() Weight {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxConjunctionScorer) Iterator() index.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxConjunctionScorer) TwoPhaseIterator() TwoPhaseIterator {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxConjunctionScorer) AdvanceShallow(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxConjunctionScorer) GetMaxScore(upTo int) (float64, error) {
	//TODO implement me
	panic("implement me")
}
