package search

import (
	"fmt"
	"github.com/geange/lucene-go/core/index"
	"math"
)

var _ Query = &BoostQuery{}

// BoostQuery
// A Query wrapper that allows to give a boost to the wrapped query.
// Boost values that are less than one will give less importance to this query compared to other ones
// while values that are greater than one will give more importance to the scores returned by this query.
// More complex boosts can be applied by using FunctionScoreQuery in the lucene-queries module
type BoostQuery struct {
	query Query
	boost float64
}

func NewBoostQuery(query Query, boost float64) (*BoostQuery, error) {
	if boost >= math.MaxInt32 || boost < 0 {
		return nil, fmt.Errorf("boost must be a positive float, got %f", boost)
	}
	return &BoostQuery{query: query, boost: boost}, nil
}

func (b *BoostQuery) String(field string) string {
	return fmt.Sprintf("(%s)^%f", b.query.String(field), b.boost)
}

func (b *BoostQuery) CreateWeight(searcher *IndexSearcher, scoreMode *ScoreMode, boost float64) (Weight, error) {
	return b.query.CreateWeight(searcher, scoreMode, b.boost*boost)
}

func (b *BoostQuery) Rewrite(reader index.Reader) (Query, error) {
	rewritten, err := b.query.Rewrite(reader)
	if err != nil {
		return nil, err
	}

	if in, ok := rewritten.(*BoostQuery); ok {
		return NewBoostQuery(b.query, b.boost*in.boost)
	}

	if _, ok := rewritten.(*ConstantScoreQuery); b.boost == 0 && !ok {
		return NewBoostQuery(NewConstantScoreQuery(rewritten), 0)
	}

	if b.query != rewritten {
		return NewBoostQuery(rewritten, b.boost)
	}

	return b, nil
}

func (b *BoostQuery) Visit(visitor QueryVisitor) error {
	return b.query.Visit(visitor.GetSubVisitor(OccurMust, b))
}

func (b *BoostQuery) GetQuery() Query {
	return b.query
}

func (b *BoostQuery) GetBoost() float64 {
	return b.boost
}
