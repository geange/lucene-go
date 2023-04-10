package search

// TopDocs
// Represents hits returned by IndexSearcher.search(Query, int).
type TopDocs struct {
	// The total number of hits for the query.
	TotalHits *TotalHits

	// The top hits for the query.
	ScoreDocs []ScoreDoc
}

// NewTopDocs Constructs a TopDocs.
func NewTopDocs(totalHits *TotalHits, scoreDocs []ScoreDoc) *TopDocs {
	return &TopDocs{TotalHits: totalHits, ScoreDocs: scoreDocs}
}
