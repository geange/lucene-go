package search

import (
	"errors"
	"fmt"
	"io"

	"github.com/geange/gods-generic/sets/treeset"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

//var _ Weight = &ConstantScoreWeight{}

type ConstantScoreWeight struct {
	*BaseWeight

	score float64
}

func (c *ConstantScoreWeight) ExtractTerms(terms *treeset.Set[index.Term]) error {
	return nil
}

func NewConstantScoreWeight(score float64, query index.Query, spi WeightScorer) *ConstantScoreWeight {
	weight := &ConstantScoreWeight{score: score}
	weight.BaseWeight = NewBaseWeight(query, spi)
	return weight
}

func (c *ConstantScoreWeight) Explain(ctx index.LeafReaderContext, doc int) (types.Explanation, error) {
	s, err := c.scorer.Scorer(ctx)
	if err != nil {
		return nil, err
	}
	exists := false
	if s != nil {
		twoPhase := s.TwoPhaseIterator()
		if twoPhase == nil {
			advance, err := s.Iterator().Advance(nil, doc)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					return nil, err
				}
			} else {
				exists = advance == doc
			}
		} else {
			isMatch, err := twoPhase.Matches()
			if err != nil {
				return nil, err
			}

			advance, err := twoPhase.Approximation().Advance(nil, doc)
			if err != nil {
				return nil, err
			}
			exists = (advance == doc) && isMatch
		}

	}

	if exists {
		return types.ExplanationMatch(c.score, c.GetQuery().String("")), nil
	}
	return types.ExplanationNoMatch(c.GetQuery().String("") + fmt.Sprintf(" doesn't match id %d", doc)), nil
}

func (c *ConstantScoreWeight) Score() float64 {
	return c.score
}
