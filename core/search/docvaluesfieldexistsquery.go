package search

import "github.com/geange/lucene-go/core/interface/index"

var _ index.Query = &DocValuesFieldExistsQuery{}

// DocValuesFieldExistsQuery
// A Query that matches documents that have a value for a given field as reported by doc values iterators.
type DocValuesFieldExistsQuery struct {
}

func (d *DocValuesFieldExistsQuery) CreateWeight(searcher index.IndexSearcher, scoreMode index.ScoreMode, boost float64) (index.Weight, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesFieldExistsQuery) Rewrite(reader index.IndexReader) (index.Query, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesFieldExistsQuery) Visit(visitor index.QueryVisitor) error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesFieldExistsQuery) String(field string) string {
	//TODO implement me
	panic("implement me")
}
