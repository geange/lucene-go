package search

import "github.com/geange/lucene-go/core/index"

var _ Query = &ConstantScoreQuery{}

// ConstantScoreQuery
// A query that wraps another query and simply returns a constant score equal to 1 for every document
// that matches the query. It therefore simply strips of all scores and always returns 1.
type ConstantScoreQuery struct {
	query Query
}

func NewConstantScoreQuery(query Query) *ConstantScoreQuery {
	return &ConstantScoreQuery{query: query}
}

func (c *ConstantScoreQuery) String(field string) string {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreQuery) CreateWeight(searcher *IndexSearcher, scoreMode *ScoreMode, boost float64) (Weight, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreQuery) Rewrite(reader index.Reader) (Query, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreQuery) Visit(visitor QueryVisitor) (err error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConstantScoreQuery) GetQuery() Query {
	return c.query
}
