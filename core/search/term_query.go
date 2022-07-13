package search

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search/similarities"
)

// TermQuery A Query that matches documents containing a term. This may be combined with other terms with a BooleanQuery.
type TermQuery struct {
	term               *index.Term
	perReaderTermState index.TermStates
}

func (t *TermQuery) getTerm() *index.Term {
	return t.term
}

func (t *TermQuery) CreateWeight(searcher *IndexSearcher, scoreMode ScoreMode, boost float64) (Weight, error) {
	//context := searcher.GetTopReaderContext()
	termState := t.perReaderTermState
	// TODO: fix it
	return NewTermWeight(searcher, scoreMode, boost, termState), nil
}

func (t *TermQuery) Rewrite(reader *index.IndexReader) (Query, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TermQuery) Visit(visitor QueryVisitor) {
	//if visitor.
}

type TermWeight struct {
	similarity similarities.Similarity
	simScorer  similarities.SimScorer
	termStates index.TermStates
	scoreMode  ScoreMode
}

func NewTermWeight(searcher *IndexSearcher, scoreMode ScoreMode, boost float64, termStates index.TermStates) *TermWeight {
	panic("")
}

func (t *TermWeight) Matches(context *index.LeafReaderContext, doc int) (Matches, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TermWeight) Match(value interface{}, description string, details []Explanation) (*Explanation, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TermWeight) NoMatch(value interface{}, description string, details []Explanation) (*Explanation, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TermWeight) IsMatch() bool {
	//TODO implement me
	panic("implement me")
}

func (t *TermWeight) GetValue() any {
	//TODO implement me
	panic("implement me")
}

func (t *TermWeight) GetDescription() string {
	//TODO implement me
	panic("implement me")
}

func (t *TermWeight) GetDetails() []Explanation {
	//TODO implement me
	panic("implement me")
}
