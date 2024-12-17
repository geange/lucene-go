package search

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/geange/gods-generic/sets/treeset"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.Query = &TermQuery{}

// TermQuery
// A Query that matches documents containing a term.
// This may be combined with other terms with a BooleanQuery.
type TermQuery struct {
	term               index.Term
	perReaderTermState *coreIndex.TermStates
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

func NewTermQuery(term index.Term) *TermQuery {
	return &TermQuery{
		term:               term,
		perReaderTermState: nil,
	}
}

func (t *TermQuery) SetStates(states *coreIndex.TermStates) *TermQuery {
	t.perReaderTermState = states
	return t
}

// NewTermQueryV1
// Expert: constructs a TermQuery that will use the provided docFreq instead of looking up
// the docFreq against the searcher.
func NewTermQueryV1(term index.Term, states *coreIndex.TermStates) *TermQuery {
	return &TermQuery{
		term:               term,
		perReaderTermState: states,
	}
}

func (t *TermQuery) GetTerm() index.Term {
	return t.term
}

func (t *TermQuery) CreateWeight(searcher index.IndexSearcher, scoreMode index.ScoreMode, boost float64) (index.Weight, error) {
	readerContext := searcher.GetTopReaderContext()

	var termState *coreIndex.TermStates
	var err error
	if t.perReaderTermState == nil || !t.perReaderTermState.WasBuiltFor(readerContext) {
		termState, err = coreIndex.BuildTermStates(readerContext, t.term, scoreMode.NeedsScores())
		if err != nil {
			return nil, err
		}
	} else {
		termState = t.perReaderTermState
	}

	return t.NewTermWeight(searcher, scoreMode, boost, termState)
}

func (t *TermQuery) Rewrite(reader index.IndexReader) (index.Query, error) {
	return t, nil
}

func (t *TermQuery) Visit(visitor index.QueryVisitor) error {
	if visitor.AcceptField(t.term.Field()) {
		visitor.ConsumeTerms(t, t.term)
	}
	return nil
}

var _ index.Weight = &TermWeight{}

type TermWeight struct {
	*BaseWeight
	*TermQuery

	similarity index.Similarity
	simScorer  index.SimScorer
	termStates *coreIndex.TermStates
	scoreMode  index.ScoreMode
}

func (t *TermWeight) IsCacheable(ctx index.LeafReaderContext) bool {
	return true
}

func (t *TermWeight) ExtractTerms(terms *treeset.Set[index.Term]) error {
	terms.Add(t.GetTerm())
	return nil
}

func (t *TermWeight) Explain(context index.LeafReaderContext, doc int) (types.Explanation, error) {
	scorer, err := t.Scorer(context)
	if err != nil {
		return nil, err
	}
	if scorer == nil {
		return nil, errors.New("no matching term")
	}

	tScorer, ok := scorer.(*TermScorer)
	if !ok {
		return nil, errors.New("no matching term")
	}

	newDoc, err := tScorer.Iterator().Advance(nil, doc)
	if err != nil {
		return nil, err
	}
	if newDoc == doc {
		freq, err := tScorer.Freq()
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

func (t *TermWeight) GetQuery() index.Query {
	//TODO implement me
	panic("implement me")
}

func (t *TermWeight) Scorer(ctx index.LeafReaderContext) (index.Scorer, error) {
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
		impacts, err := termsEnum.Impacts(coreIndex.POSTINGS_ENUM_FREQS)
		if err != nil {
			return nil, err
		}
		return NewTermScorerWithImpacts(t, impacts, scorer), nil
	} else {
		flags := coreIndex.POSTINGS_ENUM_FREQS
		if !t.scoreMode.NeedsScores() {
			flags = coreIndex.POSTINGS_ENUM_NONE
		}

		postings, err := termsEnum.Postings(nil, flags)
		if err != nil {
			return nil, err
		}
		return NewTermScorerWithPostings(t, postings, scorer), nil
	}
}

// Returns a TermsEnum positioned at this weights Term or null if the term does not exist in the given context
func (t *TermWeight) getTermsEnum(readerContext index.LeafReaderContext) (index.TermsEnum, error) {
	state, err := t.termStates.Get(readerContext)
	if err != nil {
		return nil, err
	}
	if state == nil { // term is not present in that reader
		return nil, nil
	}

	terms, err := readerContext.LeafReader().Terms(t.term.Field())
	if err != nil {
		return nil, err
	}

	termsEnum, err := terms.Iterator()
	if err != nil {
		return nil, err
	}

	if err := termsEnum.SeekExactExpert(nil, t.term.Bytes(), state); err != nil {
		return nil, err
	}

	return termsEnum, nil
}

func (t *TermQuery) NewTermWeight(searcher index.IndexSearcher, scoreMode index.ScoreMode,
	boost float64, termStates *coreIndex.TermStates) (*TermWeight, error) {
	weight := &TermWeight{
		similarity: searcher.GetSimilarity(),
		termStates: termStates,
		scoreMode:  scoreMode,
		TermQuery:  t,
	}

	weight.BaseWeight = NewBaseWeight(weight, weight)

	var collectionStats types.CollectionStatistics
	var termStats types.TermStatistics
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
		weight.simScorer = weight.similarity.Scorer(boost, collectionStats, []types.TermStatistics{termStats})
	}
	return weight, nil
}
