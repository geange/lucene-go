package search

import (
	"errors"
	"github.com/geange/lucene-go/core/util"
	"math"
)

var _ ScorerSupplier = &Boolean2ScorerSupplier{}

type Boolean2ScorerSupplier struct {
	weight         Weight
	subs           map[Occur][]ScorerSupplier
	scoreMode      *ScoreMode
	minShouldMatch int
	cost           int64
}

func NewBoolean2ScorerSupplier(weight Weight, subs map[Occur][]ScorerSupplier,
	scoreMode *ScoreMode, minShouldMatch int) (*Boolean2ScorerSupplier, error) {

	if minShouldMatch < 0 {
		return nil, errors.New("minShouldMatch must be positive")
	}

	if minShouldMatch != 0 && minShouldMatch >= len(subs[SHOULD]) {
		return nil, errors.New("minShouldMatch must be strictly less than the number of SHOULD clauses")
	}

	if scoreMode.NeedsScores() == false && minShouldMatch == 0 && len(subs[SHOULD]) > 0 &&
		len(subs[MUST])+len(subs[FILTER]) > 0 {
		return nil, errors.New("cannot pass purely optional clauses if scores are not needed")
	}

	if len(subs[SHOULD])+len(subs[MUST])+len(subs[FILTER]) == 0 {
		return nil, errors.New("there should be at least one positive clause")
	}

	return &Boolean2ScorerSupplier{
		weight:         weight,
		subs:           subs,
		scoreMode:      scoreMode,
		minShouldMatch: minShouldMatch,
	}, nil
}

func (b *Boolean2ScorerSupplier) Get(leadCost int64) (Scorer, error) {
	scorer, err := b.getInternal(leadCost)
	if err != nil {
		return nil, err
	}

	if b.scoreMode == TOP_SCORES && len(b.subs[SHOULD]) == 0 && len(b.subs[MUST]) == 0 {
		// no scoring clauses but scores are needed so we wrap the scorer in
		// a constant score in order to allow early termination
		if scorer.TwoPhaseIterator() != nil {
			return NewConstantScoreScorerV1(b.weight, 0, b.scoreMode, scorer.TwoPhaseIterator())
		} else {
			return NewConstantScoreScorer(b.weight, 0, b.scoreMode, scorer.Iterator())
		}
	}
	return scorer, nil
}

func (b *Boolean2ScorerSupplier) Cost() int64 {
	if b.cost == -1 {
		b.cost = b.computeCost()
	}
	return b.cost

}

func (b *Boolean2ScorerSupplier) getInternal(leadCost int64) (Scorer, error) {
	panic("")
}

var _ Scorer = &filterScorer{}

type filterScorer struct {
	*FilterScorer
}

func (f *filterScorer) Score() (float64, error) {
	return 0, nil
}

func (f *filterScorer) GetMaxScore(upTo int) (float64, error) {
	return 0, nil
}

// Create a new scorer for the given required clauses. Note that requiredScoring is a subset of required containing required clauses that should participate in scoring.
func (b *Boolean2ScorerSupplier) req(requiredNoScoring, requiredScoring []ScorerSupplier,
	leadCost int64) (Scorer, error) {

	if len(requiredNoScoring)+len(requiredScoring) == 1 {
		var req Scorer
		var err error

		if len(requiredNoScoring) == 0 {
			req, err = requiredScoring[0].Get(leadCost)
			if err != nil {
				return nil, err
			}
		} else {
			req, err = requiredNoScoring[0].Get(leadCost)
			if err != nil {
				return nil, err
			}
		}

		if len(requiredScoring) == 0 {
			// Scores are needed but we only have a filter clause
			// BooleanWeight expects that calling score() is ok so we need to wrap
			// to prevent score() from being propagated
			return &filterScorer{newFilterScorer(req)}, nil
		}

		return req, nil
	}
	requiredScorers := make([]Scorer, 0)
	scoringScorers := make([]Scorer, 0)
	for _, s := range requiredNoScoring {
		scorer, err := s.Get(leadCost)
		if err != nil {
			return nil, err
		}
		requiredScorers = append(requiredScorers, scorer)
	}

	for _, s := range requiredScoring {
		scorer, err := s.Get(leadCost)
		if err != nil {
			return nil, err
		}
		scoringScorers = append(scoringScorers, scorer)
	}

	if b.scoreMode == TOP_SCORES && len(scoringScorers) > 1 {
		blockMaxScorer := NewBlockMaxConjunctionScorer(b.weight, scoringScorers)
		if len(requiredScorers) == 0 {
			return blockMaxScorer, nil
		}
		scoringScorers = []Scorer{blockMaxScorer}
	}
	requiredScorers = append(requiredScorers, scoringScorers...)
	return NewConjunctionScorer(b.weight, requiredScorers, scoringScorers)
}

func (b *Boolean2ScorerSupplier) excl(main Scorer, prohibited []ScorerSupplier,
	leadCost int64) (Scorer, error) {
	if len(prohibited) == 0 {
		return main, nil
	}
	opt, err := b.opt(prohibited, 1, COMPLETE_NO_SCORES, leadCost)
	if err != nil {
		return nil, err
	}
	return NewReqExclScorer(main, opt), nil
}

func (b *Boolean2ScorerSupplier) opt(optional []ScorerSupplier, minShouldMatch int,
	scoreMode *ScoreMode, leadCost int64) (Scorer, error) {

	if len(optional) == 1 {
		return optional[0].Get(leadCost)
	}

	optionalScorers := make([]Scorer, 0)
	for _, v := range optional {
		scorer, err := v.Get(leadCost)
		if err != nil {
			return nil, err
		}
		optionalScorers = append(optionalScorers, scorer)
	}

	// Technically speaking, WANDScorer should be able to handle the following 3 conditions now
	// 1. Any ScoreMode (with scoring or not)
	// 2. Any minCompetitiveScore ( >= 0 )
	// 3. Any minShouldMatch ( >= 0 )
	//
	// However, as WANDScorer uses more complex algorithm and data structure, we would like to
	// still use DisjunctionSumScorer to handle exhaustive pure disjunctions, which may be faster
	if scoreMode == TOP_SCORES || minShouldMatch > 1 {
		return newWANDScorer(b.weight, optionalScorers, minShouldMatch, scoreMode)
	}
	return newDisjunctionScorer(b.weight, optionalScorers, scoreMode)
}

func (b *Boolean2ScorerSupplier) computeCost() int64 {
	minRequiredCost := int64(math.MaxInt64)

	for _, supplier := range b.subs[MUST] {
		cost := supplier.Cost()
		if cost < minRequiredCost {
			minRequiredCost = cost
		}
	}

	for _, supplier := range b.subs[FILTER] {
		cost := supplier.Cost()
		if cost < minRequiredCost {
			minRequiredCost = cost
		}
	}

	if b.minShouldMatch == 0 && minRequiredCost != int64(math.MaxInt64) {
		return minRequiredCost
	} else {
		optionalScorers := b.subs[SHOULD]
		costs := make([]int64, 0, len(optionalScorers))
		for _, scorer := range optionalScorers {
			costs = append(costs, scorer.Cost())
		}
		shouldCost := costWithMinShouldMatch(costs, len(optionalScorers), b.minShouldMatch)
		return util.Min(minRequiredCost, shouldCost)
	}
}
