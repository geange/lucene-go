package search

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
	"reflect"
)

var _ Query = &TermQuery{}

// TermQuery A Query that matches documents containing a term.
// This may be combined with other terms with a BooleanQuery.
type TermQuery struct {
	term               *index.Term
	perReaderTermState *index.TermStates
}

func (t *TermQuery) String(field string) string {
	buf := new(bytes.Buffer)
	if t.term.Field() != field {
		buf.WriteString(t.term.Field())
		buf.WriteString(":")
	}
	buf.WriteString(t.term.Text())
	return buf.String()
}

func NewTermQuery(term *index.Term) *TermQuery {
	return &TermQuery{
		term:               term,
		perReaderTermState: nil,
	}
}

// NewTermQueryV1
// Expert: constructs a TermQuery that will use the provided docFreq instead of looking up
// the docFreq against the searcher.
func NewTermQueryV1(term *index.Term, states *index.TermStates) *TermQuery {
	return &TermQuery{
		term:               term,
		perReaderTermState: states,
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

func (t *TermWeight) Explain(context *index.LeafReaderContext, doc int) (*types.Explanation, error) {
	scorer, err := t.Scorer(context)
	if err != nil {
		return nil, err
	}
	if scorer == nil {
		return nil, errors.New("no matching term")
	}

	tscorer, ok := scorer.(*TermScorer)
	if !ok {
		return nil, errors.New("no matching term")
	}

	newDoc, err := tscorer.Iterator().Advance(doc)
	if err != nil {
		return nil, err
	}
	if newDoc == doc {
		freq, err := tscorer.Freq()
		if err != nil {
			return nil, err
		}
		docScorer, err := NewLeafSimScorer(t.simScorer, context.Reader().(index.LeafReader), t.term.Field(), true)
		if err != nil {
			return nil, err
		}

		freqExplanation := types.ExplanationMatch(freq, "freq, occurrences of term within document")
		scoreExplanation, err := docScorer.Explain(doc, freqExplanation)
		if err != nil {
			return nil, err
		}
		return types.ExplanationMatch(scoreExplanation.GetValue().(float64),
			fmt.Sprintf(`weight(%s in %d) [%s]`, t.GetQuery(), doc, reflect.TypeOf(t.similarity).Name()),
			scoreExplanation), nil
	}

	return nil, errors.New("no matching term")
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
