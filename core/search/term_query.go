package search

import (
	"github.com/geange/lucene-go/core/index"
)

// TermQuery A Query that matches documents containing a term. This may be combined with other terms with a BooleanQuery.
type TermQuery struct {
	term               *index.Term
	perReaderTermState *index.TermStates
}

func NewTermQuery(term *index.Term) *TermQuery {
	return &TermQuery{
		term:               term,
		perReaderTermState: nil,
	}
}

func (t *TermQuery) GetTerm() *index.Term {
	return t.term
}

func (t *TermQuery) CreateWeight(searcher *IndexSearcher, scoreMode *ScoreMode, boost float64) (Weight, error) {
	context := searcher.GetTopReaderContext()

	var termState *index.TermStates
	var err error
	if t.perReaderTermState == nil || !t.perReaderTermState.WasBuiltFor(context) {
		termState, err = index.BuildTermStates(context, t.term, scoreMode.NeedsScores())
		if err != nil {
			return nil, err
		}
	} else {
		termState = t.perReaderTermState
	}

	return NewTermWeight(searcher, scoreMode, boost, termState), nil
}

func (t *TermQuery) Rewrite(reader index.IndexReader) (Query, error) {
	return t, nil
}

func (t *TermQuery) Visit(visitor QueryVisitor) {
	if visitor.AcceptField(t.term.Field()) {
		visitor.ConsumeTerms(t, t.term)
	}
}

var _ Weight = &TermWeight{}

type TermWeight struct {
	*WeightImp

	similarity Similarity
	simScorer  SimScorer
	termStates index.TermStates
	scoreMode  ScoreMode
}

func (t TermWeight) Explain(ctx *index.LeafReaderContext, doc int) (*Explanation, error) {
	//TODO implement me
	panic("implement me")
}

func (t TermWeight) GetQuery() Query {
	//TODO implement me
	panic("implement me")
}

func NewTermWeight(searcher *IndexSearcher, scoreMode *ScoreMode, boost float64, termStates *index.TermStates) *TermWeight {
	//newWeightImp()
	panic("implement me")
}
