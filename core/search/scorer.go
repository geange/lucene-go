package search

import (
	"github.com/geange/lucene-go/core/interface/search"
	"github.com/geange/lucene-go/core/types"
)

type BaseScorer struct {
	*BaseScorable

	weight search.Weight
}

func NewScorer(weight search.Weight) *BaseScorer {
	return &BaseScorer{weight: weight}
}

func (s *BaseScorer) GetWeight() search.Weight {
	return s.weight
}

func (s *BaseScorer) TwoPhaseIterator() search.TwoPhaseIterator {
	return nil
}

func (s *BaseScorer) AdvanceShallow(target int) (int, error) {
	return types.NO_MORE_DOCS, nil
}
