package search

import (
	"github.com/geange/lucene-go/core/index"
)

var _ Query = &MatchAllDocsQuery{}

type MatchAllDocsQuery struct {
}

func (m *MatchAllDocsQuery) String(field string) string {
	//TODO implement me
	panic("implement me")
}

func (m *MatchAllDocsQuery) CreateWeight(searcher *IndexSearcher, scoreMode *ScoreMode, boost float64) (Weight, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MatchAllDocsQuery) Rewrite(reader index.IndexReader) (Query, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MatchAllDocsQuery) Visit(visitor QueryVisitor) {
	//TODO implement me
	panic("implement me")
}
