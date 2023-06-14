package search

import (
	"fmt"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
	"io"
)

//var _ Weight = &ConstantScoreWeight{}

type ConstantScoreWeight struct {
	score float64

	*WeightDefault

	//FnScorer func(ctx *index.LeafReaderContext) (Scorer, error)
}

func NewConstantScoreWeight(score float64, query Query, spi WeightSPI) *ConstantScoreWeight {
	weight := &ConstantScoreWeight{score: score}
	weight.WeightDefault = NewWeight(query, spi)
	return weight
}

func (c *ConstantScoreWeight) Explain(ctx *index.LeafReaderContext, doc int) (*types.Explanation, error) {
	s, err := c.Scorer(ctx)
	if err != nil {
		return nil, err
	}
	exists := false
	if s != nil {
		twoPhase := s.TwoPhaseIterator()
		if twoPhase == nil {
			advance, err := s.Iterator().Advance(doc)
			if err == nil {
				exists = advance == doc
			} else if err != nil && err != io.EOF {
				return nil, err
			}
		} else {
			matches, err := twoPhase.Matches()
			if err != nil {
				return nil, err
			}

			advance, err := twoPhase.Approximation().Advance(doc)
			if err == nil {
				exists = (advance == doc) && matches
			}
			if err != nil {
				return nil, err
			}
		}

	}

	if exists {
		return types.ExplanationMatch(c.score, c.GetQuery().String("")), nil
	}
	return types.ExplanationNoMatch(c.GetQuery().String("") + fmt.Sprintf(" doesn't match id %d", doc)), nil
}
