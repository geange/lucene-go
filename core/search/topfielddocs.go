package search

import "github.com/geange/lucene-go/core/index"

type TopFieldDocs struct {
	*TopDocsDefault

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
		TopDocsDefault: NewTopDocs(totalHits, scoreDocs),
		fields:         fields,
	}
}

func (t *TopFieldDocs) GetFields() []index.SortField {
	return t.fields
}
