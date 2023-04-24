package search

type ScoreDoc interface {
	GetScore() float64
	SetScore(score float64)
	GetDoc() int
	SetDoc(doc int)
	GetShardIndex() int
	SetShardIndex(shardIndex int)
}

var _ ScoreDoc = &ScoreDocDefault{}

// ScoreDocDefault
// Holds one hit in TopDocs.
type ScoreDocDefault struct {
	// The score of this document for the query.
	score float64

	// A hit document's number.
	// See Also: IndexSearcher.doc(int)
	doc int

	// Only set by TopDocs.merge
	shardIndex int
}

func (s *ScoreDocDefault) GetScore() float64 {
	return s.score
}

func (s *ScoreDocDefault) SetScore(score float64) {
	s.score = score
}

func (s *ScoreDocDefault) GetDoc() int {
	return s.doc
}

func (s *ScoreDocDefault) SetDoc(doc int) {
	s.doc = doc
}

func (s *ScoreDocDefault) GetShardIndex() int {
	return s.shardIndex
}

func (s *ScoreDocDefault) SetShardIndex(shardIndex int) {
	s.shardIndex = shardIndex
}

func NewScoreDoc(doc int, score float64) *ScoreDocDefault {
	return &ScoreDocDefault{score: score, doc: doc, shardIndex: -1}
}

func NewScoreDocV1(score float64, doc int, shardIndex int) *ScoreDocDefault {
	return &ScoreDocDefault{score: score, doc: doc, shardIndex: shardIndex}
}
