package search

import (
	"github.com/geange/lucene-go/core/interface/index"
)

type TopFieldDocs struct {
	*BaseTopDocs

	fields []index.SortField
}

// NewTopFieldDocs
// Creates one of these objects.
// Params:
//
//	totalHits – Total number of hits for the query.
//	scoreDocs – The top hits for the query.
//	fields – The sort criteria used to find the top hits.
func NewTopFieldDocs(totalHits *TotalHits, scoreDocs []ScoreDoc, fields []index.SortField) *TopFieldDocs {
	return &TopFieldDocs{
		BaseTopDocs: NewTopDocs(totalHits, scoreDocs),
		fields:      fields,
	}
}

func (t *TopFieldDocs) GetFields() []index.SortField {
	return t.fields
}
