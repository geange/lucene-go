package search

import (
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.Query = &MatchNoDocsQuery{}

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

func (m *MatchNoDocsQuery) CreateWeight(searcher index.IndexSearcher, scoreMode index.ScoreMode, boost float64) (index.Weight, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MatchNoDocsQuery) Rewrite(reader index.IndexReader) (index.Query, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MatchNoDocsQuery) Visit(visitor index.QueryVisitor) (err error) {
	//TODO implement me
	panic("implement me")
}
