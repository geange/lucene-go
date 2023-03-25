package search

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
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

	return t.NewTermWeight(searcher, scoreMode, boost, termState)
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
	*WeightDefault

	similarity index.Similarity
	simScorer  index.SimScorer
	termStates *index.TermStates
	scoreMode  *ScoreMode

	*TermQuery
}

func (t *TermWeight) Explain(ctx *index.LeafReaderContext, doc int) (*Explanation, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TermWeight) GetQuery() Query {
	//TODO implement me
	panic("implement me")
}

func (t *TermWeight) Scorer(ctx *index.LeafReaderContext) (Scorer, error) {
	termsEnum, err := t.getTermsEnum(ctx)
	if err != nil {
		return nil, err
	}

	if termsEnum == nil {
		return nil, nil
	}

	scorer, err := NewLeafSimScorer(t.simScorer, ctx.LeafReader(), t.term.Field(), t.scoreMode.NeedsScores())
	if err != nil {
		return nil, err
	}

	if t.scoreMode == TOP_SCORES {
		impacts, err := termsEnum.Impacts(index.POSTINGS_ENUM_FREQS)
		if err != nil {
			return nil, err
		}
		return NewTermScorerWithImpacts(t, impacts, scorer), nil
	} else {
		flags := index.POSTINGS_ENUM_FREQS
		if !t.scoreMode.NeedsScores() {
			flags = index.POSTINGS_ENUM_NONE
		}

		postings, err := termsEnum.Postings(nil, flags)
		if err != nil {
			return nil, err
		}
		return NewTermScorerWithPostings(t, postings, scorer), nil
	}
}

// Returns a TermsEnum positioned at this weights Term or null if the term does not exist in the given context
func (t *TermWeight) getTermsEnum(context *index.LeafReaderContext) (index.TermsEnum, error) {
	state, err := t.termStates.Get(context)
	if err != nil {
		return nil, err
	}
	if state == nil { // term is not present in that reader
		return nil, nil
	}
	terms, err := context.LeafReader().Terms(t.term.Field())
	if err != nil {
		return nil, err
	}
	termsEnum, err := terms.Iterator()
	if err != nil {
		return nil, err
	}

	err = termsEnum.SeekExactExpert(t.term.Bytes(), state)
	if err != nil {
		return nil, err
	}
	return termsEnum, nil
}

func (t *TermQuery) NewTermWeight(searcher *IndexSearcher, scoreMode *ScoreMode,
	boost float64, termStates *index.TermStates) (*TermWeight, error) {
	weight := &TermWeight{
		similarity: searcher.GetSimilarity(),
		termStates: termStates,
		scoreMode:  scoreMode,
		TermQuery:  t,
	}

	weight.WeightDefault = NewWeight(weight, weight)

	var collectionStats *types.CollectionStatistics
	var termStats *types.TermStatistics
	var err error
	if scoreMode.NeedsScores() {
		collectionStats, err = searcher.CollectionStatistics(t.term.Field())
		if err != nil {
			return nil, err
		}

		freq, err := termStates.DocFreq()
		if err != nil {
			return nil, err
		}

		if freq > 0 {
			docFreq, err := termStates.DocFreq()
			if err != nil {
				return nil, err
			}
			totalTermFreq, err := termStates.TotalTermFreq()
			if err != nil {
				return nil, err
			}

			termStats, err = searcher.TermStatistics(t.term, docFreq, int(totalTermFreq))
			if err != nil {
				return nil, err
			}
		}
	} else {
		var err error
		collectionStats, err = types.NewCollectionStatistics(t.term.Field(), 1, 1, 1, 1)
		if err != nil {
			return nil, err
		}
		termStats, err = types.NewTermStatistics(t.term.Bytes(), 1, 1)
		if err != nil {
			return nil, err
		}
	}

	if termStats == nil {
		weight.simScorer = nil
	} else {
		weight.simScorer = weight.similarity.Scorer(boost, collectionStats, []types.TermStatistics{*termStats})
	}
	return weight, nil
}
