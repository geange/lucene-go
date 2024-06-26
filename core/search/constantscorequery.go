package search

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/interface/search"
)

var _ search.Query = &ConstantScoreQuery{}

// ConstantScoreQuery
// A query that wraps another query and simply returns a constant score equal to 1 for every document
// that matches the query. It therefore simply strips of all scores and always returns 1.
type ConstantScoreQuery struct {
	query search.Query
}

func NewConstantScoreQuery(query search.Query) *ConstantScoreQuery {
	return &ConstantScoreQuery{query: query}
}

func (c *ConstantScoreQuery) String(field string) string {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreQuery) CreateWeight(searcher search.IndexSearcher, scoreMode search.ScoreMode, boost float64) (search.Weight, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreQuery) Rewrite(reader index.IndexReader) (search.Query, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreQuery) Visit(visitor search.QueryVisitor) (err error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreQuery) GetQuery() search.Query {
	return c.query
}
