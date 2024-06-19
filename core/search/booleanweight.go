package search

import (
	"context"
	"github.com/geange/gods-generic/sets/treeset"
	index2 "github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/interface/search"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"math"
)

var _ search.Weight = &BooleanWeight{}

func NewBooleanWeight(query *BooleanQuery, searcher search.IndexSearcher,
	scoreMode search.ScoreMode, boost float64) (*BooleanWeight, error) {
	weight := &BooleanWeight{
		similarity:      searcher.GetSimilarity(),
		query:           query,
		weightedClauses: make([]*weightedBooleanClause, 0),
		scoreMode:       scoreMode,
	}
	weight.BaseWeight = NewBaseWeight(query, weight)

	for _, c := range query.Clauses() {
		mode := COMPLETE_NO_SCORES
		if c.IsScoring() {
			mode = scoreMode
		}

		w, err := searcher.CreateWeight(c.GetQuery(), mode, boost)
		if err != nil {
			return nil, err
		}
		weight.weightedClauses = append(weight.weightedClauses, newWeightedBooleanClause(c, w))
	}

	weight.query = query

	return weight, nil
}

// BooleanWeight
// Expert: the Weight for BooleanQuery, used to normalize, score and explain these queries.
type BooleanWeight struct {
	*BaseWeight

	similarity      index2.Similarity
	query           *BooleanQuery
	weightedClauses []*weightedBooleanClause
	scoreMode       search.ScoreMode
}

func (b *BooleanWeight) ExtractTerms(terms *treeset.Set[index2.Term]) error {
	for _, wc := range b.weightedClauses {
		if wc.clause.IsScoring() || (b.scoreMode.NeedsScores() == false && wc.clause.IsProhibited() == false) {
			if err := wc.weight.ExtractTerms(terms); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *BooleanWeight) Explain(ctx index2.LeafReaderContext, doc int) (*types.Explanation, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BooleanWeight) Matches(context index2.LeafReaderContext, doc int) (search.Matches, error) {
	minShouldMatch := b.query.GetMinimumNumberShouldMatch()
	matchValues := make([]search.Matches, 0)
	shouldMatchCount := 0
	for _, wc := range b.weightedClauses {
		w := wc.weight
		bc := wc.clause
		m, err := w.Matches(context, doc)
		if err != nil {
			return nil, err
		}
		if bc.IsProhibited() {
			if m != nil {
				return nil, nil
			}
		}
		if bc.IsRequired() {
			if m == nil {
				return nil, nil
			}
			matchValues = append(matchValues, m)
		}
		if bc.GetOccur() == search.OccurShould {
			if m != nil {
				matchValues = append(matchValues, m)
				shouldMatchCount++
			}
		}
	}
	if shouldMatchCount < minShouldMatch {
		return nil, nil
	}

	return MatchesFromSubMatches(matchValues)
}

func disableScoring(scorer search.BulkScorer) search.BulkScorer {
	return &BulkScorerAnon{
		FnScoreRange: func(collector search.LeafCollector, acceptDocs util.Bits, min, max int) (int, error) {
			fake := NewScoreAndDoc()
			noScoreCollector := &LeafCollectorAnon{
				FnSetScorer: func(scorer search.Scorable) error {
					return collector.SetScorer(fake)
				},
				FnCollect: func(ctx context.Context, doc int) error {
					fake.doc = doc
					return collector.Collect(ctx, doc)
				},
				FnCompetitiveIterator: nil,
			}
			return scorer.ScoreRange(noScoreCollector, acceptDocs, min, max)
		},
		FnCost: func() int64 {
			return scorer.Cost()
		},
	}
}

func (*BooleanWeight) optionalBulkScorer(context index2.LeafReaderContext) (search.BulkScorer, error) {
	panic("")
}

// Return a BulkScorer for the required clauses only,
// or null if it is not applicable
func (b *BooleanWeight) requiredBulkScorer(context index2.LeafReaderContext) (search.BulkScorer, error) {
	var scorer search.BulkScorer
	var err error

	for _, wc := range b.weightedClauses {
		w := wc.weight
		c := wc.clause
		if c.IsRequired() == false {
			continue
		}
		if scorer != nil {
			// we don't have a BulkScorer for conjunctions
			return nil, nil
		}
		scorer, err = w.BulkScorer(context)
		if err != nil {
			return nil, err
		}

		if scorer == nil {
			// no matches
			return nil, nil
		}
		if c.IsScoring() == false && b.scoreMode.NeedsScores() {
			scorer = disableScoring(scorer)
		}
	}
	return scorer, nil
}

// Try to build a boolean scorer for this weight. Returns null if BooleanScorer cannot be used.
func (b *BooleanWeight) booleanScorer(context index2.LeafReaderContext) (search.BulkScorer, error) {
	numOptionalClauses := len(b.query.GetClauses(search.OccurShould))
	numRequiredClauses := len(b.query.GetClauses(search.OccurMust)) + len(b.query.GetClauses(search.OccurFilter))

	var positiveScorer search.BulkScorer
	var err error
	if numRequiredClauses == 0 {
		positiveScorer, err = b.optionalBulkScorer(context)
		if err != nil {
			return nil, err
		}
		if positiveScorer == nil {
			return nil, nil
		}

		// TODO: what is the right heuristic here?
		var costThreshold int64
		if b.query.GetMinimumNumberShouldMatch() <= 1 {
			// when all clauses are optional, use BooleanScorer aggressively
			// TODO: is there actually a threshold under which we should rather
			// use the regular scorer?
			costThreshold = -1
		} else {
			// when a minimum number of clauses should match, BooleanScorer is
			// going to score all windows that have at least minNrShouldMatch
			// matches in the window. But there is no way to know if there is
			// an intersection (all clauses might match a different doc ID and
			// there will be no matches in the end) so we should only use
			// BooleanScorer if matches are very dense
			costThreshold = int64(context.Reader().MaxDoc() / 3)
		}

		if positiveScorer.Cost() < costThreshold {
			return nil, nil
		}

	} else if numRequiredClauses == 1 && numOptionalClauses == 0 && b.query.GetMinimumNumberShouldMatch() == 0 {
		positiveScorer, err = b.requiredBulkScorer(context)
		if err != nil {
			return nil, err
		}
	} else {
		// TODO: there are some cases where BooleanScorer
		// would handle conjunctions faster than
		// BooleanScorer2...
		return nil, nil
	}

	if positiveScorer == nil {
		return nil, nil
	}

	prohibited := make([]search.Scorer, 0)
	for _, wc := range b.weightedClauses {
		//w := wc.weight
		//c := wc.clause
		if wc.clause.IsProhibited() {
			scorer, err := wc.weight.Scorer(context)
			if err != nil {
				return nil, err
			}
			if scorer != nil {
				prohibited = append(prohibited, scorer)
			}
		}
	}

	if len(prohibited) == 0 {
		return positiveScorer, nil
	} else {
		var prohibitedScorer search.Scorer
		if len(prohibited) == 1 {
			prohibitedScorer = prohibited[0]
		} else {
			prohibitedScorer, err = newDisjunctionScorer(b, prohibited, COMPLETE_NO_SCORES)
			if err != nil {
				return nil, err
			}
		}

		if prohibitedScorer.TwoPhaseIterator() != nil {
			// ReqExclBulkScorer can't deal efficiently with two-phased prohibited clauses
			return nil, nil
		}
		return newReqExclBulkScorer(positiveScorer, prohibitedScorer.Iterator()), nil
	}
}

func (b *BooleanWeight) BulkScorer(context index2.LeafReaderContext) (search.BulkScorer, error) {
	if b.scoreMode == TOP_SCORES {
		// If only the top docs are requested, use the default bulk scorer
		// so that we can dynamically prune non-competitive hits.
		return b.BaseWeight.BulkScorer(context)
	}
	bulkScorer, err := b.booleanScorer(context)
	if err != nil {
		return nil, err
	}
	if bulkScorer != nil {
		// bulk scoring is applicable, use it
		return bulkScorer, nil
	} else {
		// use a Scorer-based impl (BS2)
		return b.BaseWeight.BulkScorer(context)
	}
}

func (b *BooleanWeight) Scorer(ctx index2.LeafReaderContext) (search.Scorer, error) {
	supplier, err := b.ScorerSupplier(ctx)
	if err != nil {
		return nil, err
	}
	if supplier == nil {
		return nil, nil
	}
	return supplier.Get(math.MaxInt64)
}

func (b *BooleanWeight) IsCacheable(ctx index2.LeafReaderContext) bool {
	if len(b.query.Clauses()) > BOOLEAN_REWRITE_TERM_COUNT_THRESHOLD {
		// Disallow caching large boolean queries to not encourage users
		// to build large boolean queries as a workaround to the fact that
		// we disallow caching large TermInSetQueries.
		return false
	}
	for _, wc := range b.weightedClauses {
		w := wc.weight
		if w.IsCacheable(ctx) == false {
			return false
		}
	}
	return true
}

func (b *BooleanWeight) ScorerSupplier(context index2.LeafReaderContext) (search.ScorerSupplier, error) {
	minShouldMatch := b.query.GetMinimumNumberShouldMatch()

	scorers := map[search.Occur][]search.ScorerSupplier{}
	for _, occur := range search.OccurValues() {
		scorers[occur] = []search.ScorerSupplier{}
	}

	for _, wc := range b.weightedClauses {
		//w := wc.weight
		c := wc.clause
		subScorer, err := wc.weight.ScorerSupplier(context)
		if err != nil {
			return nil, err
		}
		if subScorer == nil {
			if c.IsRequired() {
				return nil, nil
			}
		} else {
			scorers[c.GetOccur()] = append(scorers[c.GetOccur()], subScorer)
		}
	}

	// scorer simplifications:

	if len(scorers[search.OccurShould]) == minShouldMatch {
		// any optional clauses are in fact required
		scorers[search.OccurMust] = append(scorers[search.OccurMust], scorers[search.OccurShould]...)
		scorers[search.OccurShould] = scorers[search.OccurShould][:0]
		minShouldMatch = 0
	}

	if len(scorers[search.OccurFilter]) == 0 && len(scorers[search.OccurMust]) == 0 && len(scorers[search.OccurShould]) == 0 {
		// no required and optional clauses.
		return nil, nil
	} else if len(scorers[search.OccurShould]) < minShouldMatch {
		return nil, nil
	}

	return NewBoolean2ScorerSupplier(b, scorers, b.scoreMode, minShouldMatch)
}

type weightedBooleanClause struct {
	clause *BooleanClause
	weight search.Weight
}

func newWeightedBooleanClause(clause *BooleanClause, weight search.Weight) *weightedBooleanClause {
	return &weightedBooleanClause{clause: clause, weight: weight}
}
