package search

import (
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.Query = &ConstantScoreQuery{}

// ConstantScoreQuery
// A query that wraps another query and simply returns a constant score equal to 1 for every document
// that matches the query. It therefore simply strips of all scores and always returns 1.
type ConstantScoreQuery struct {
	query index.Query
}

func NewConstantScoreQuery(query index.Query) *ConstantScoreQuery {
	return &ConstantScoreQuery{query: query}
}

func (c *ConstantScoreQuery) String(field string) string {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreQuery) CreateWeight(searcher index.IndexSearcher, scoreMode index.ScoreMode, boost float64) (index.Weight, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreQuery) Rewrite(reader index.IndexReader) (index.Query, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreQuery) Visit(visitor index.QueryVisitor) (err error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreQuery) GetQuery() index.Query {
	return c.query
}
