package search

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search/similarities"
)

// TermQuery A Query that matches documents containing a term. This may be combined with other terms with a BooleanQuery.
type TermQuery struct {
	term               index.Term
	perReaderTermState index.TermStates
}

func (t *TermQuery) ToString(field string) string {
	//TODO implement me
	panic("implement me")
}

func (t *TermQuery) CreateWeight(searcher *IndexSearcher, scoreMode ScoreMode, boost float64) (Weight, error) {
	panic("")
}

func (t *TermQuery) Rewrite(reader *index.IndexReader) (Query, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TermQuery) Visit(visitor QueryVisitor) {
	//TODO implement me
	panic("implement me")
}

type TermWeight struct {
	similarity similarities.Similarity
	simScorer  similarities.SimScorer
	termStates index.TermStates
	scoreMode  ScoreMode
}
