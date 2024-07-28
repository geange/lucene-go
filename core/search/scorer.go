package search

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

type BaseScorer struct {
	*BaseScorable

	weight index.Weight
}

func NewScorer(weight index.Weight) *BaseScorer {
	return &BaseScorer{weight: weight}
}

func (s *BaseScorer) GetWeight() index.Weight {
	return s.weight
}

func (s *BaseScorer) TwoPhaseIterator() index.TwoPhaseIterator {
	return nil
}

func (s *BaseScorer) AdvanceShallow(target int) (int, error) {
	return types.NO_MORE_DOCS, nil
}
