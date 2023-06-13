package search

import "errors"

var _ ScorerSupplier = &Boolean2ScorerSupplier{}

type Boolean2ScorerSupplier struct {
	weight         Weight
	subs           map[Occur][]ScorerSupplier
	scoreMode      *ScoreMode
	minShouldMatch int
	cost           int
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
	//     if (cost == -1) {
	//      cost = computeCost();
	//    }
	//    return cost;
	return 0
}

func (b *Boolean2ScorerSupplier) getInternal(leadCost int64) (Scorer, error) {
	panic("")
}

// Create a new scorer for the given required clauses. Note that requiredScoring is a subset of required containing required clauses that should participate in scoring.
func (b *Boolean2ScorerSupplier) req(requiredNoScoring, requiredScoring []ScorerSupplier,
	leadCost int64) (Scorer, error) {

	panic("")
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
