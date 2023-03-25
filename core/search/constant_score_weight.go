package search

import (
	"github.com/geange/lucene-go/core/index"
)

//var _ Weight = &ConstantScoreWeight{}

type ConstantScoreWeight struct {
	*WeightDefault
}

func (c *ConstantScoreWeight) Matches(context *index.LeafReaderContext, doc int) (Matches, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreWeight) Explain(ctx *index.LeafReaderContext, doc int) (*Explanation, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreWeight) GetQuery() Query {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreWeight) Scorer(ctx *index.LeafReaderContext) (Scorer, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreWeight) ScorerSupplier(ctx *index.LeafReaderContext) (ScorerSupplier, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreWeight) BulkScorer(ctx *index.LeafReaderContext) (BulkScorer, error) {
	//TODO implement me
	panic("implement me")
}
