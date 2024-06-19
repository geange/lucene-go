package search

import "github.com/geange/lucene-go/core/interface/search"

var _ search.ScoreDoc = &baseScoreDoc{}

// baseScoreDoc
// Holds one hit in TopDocs.
type baseScoreDoc struct {
	// The score of this document for the query.
	score float64

	// A hit document's number.
	// See Also: IndexSearcher.doc(int)
	doc int

	// Only set by TopDocs.merge
	shardIndex int
}

func (s *baseScoreDoc) GetScore() float64 {
	return s.score
}

func (s *baseScoreDoc) SetScore(score float64) {
	s.score = score
}

func (s *baseScoreDoc) GetDoc() int {
	return s.doc
}

func (s *baseScoreDoc) SetDoc(doc int) {
	s.doc = doc
}

func (s *baseScoreDoc) GetShardIndex() int {
	return s.shardIndex
}

func (s *baseScoreDoc) SetShardIndex(shardIndex int) {
	s.shardIndex = shardIndex
}

func newScoreDoc(doc int, score float64) *baseScoreDoc {
	return &baseScoreDoc{score: score, doc: doc, shardIndex: -1}
}

func newScoreDocWIthShard(score float64, doc int, shardIndex int) *baseScoreDoc {
	return &baseScoreDoc{score: score, doc: doc, shardIndex: shardIndex}
}
