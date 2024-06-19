package search

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/interface/search"
)

var _ search.Query = &MatchNoDocsQuery{}

// MatchNoDocsQuery
// A query that matches no documents.
type MatchNoDocsQuery struct {
	reason string
}

func NewMatchNoDocsQuery(reason string) *MatchNoDocsQuery {
	return &MatchNoDocsQuery{reason: reason}
}

func (m *MatchNoDocsQuery) String(field string) string {
	//TODO implement me
	panic("implement me")
}

func (m *MatchNoDocsQuery) CreateWeight(searcher search.IndexSearcher, scoreMode search.ScoreMode, boost float64) (search.Weight, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MatchNoDocsQuery) Rewrite(reader index.IndexReader) (search.Query, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MatchNoDocsQuery) Visit(visitor search.QueryVisitor) (err error) {
	//TODO implement me
	panic("implement me")
}
