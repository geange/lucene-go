package search

import (
	"github.com/geange/lucene-go/core/interface/search"
	"github.com/geange/lucene-go/core/types"
	"math"
)

var _ search.Scorer = &ReqOptSumScorer{}

// ReqOptSumScorer
// A Scorer for queries with a required part and an optional part.
// Delays skipTo() on the optional part until a score() is needed.
//
// GPT3.5:
//
// 在Lucene中，ReqOptSumScorer是一个用于布尔查询的评分器（scorer）。
// 它是由ReqScorer和OptScorer组合而成，用于计算布尔查询的相关性得分。
//
// ReqScorer（必须匹配评分器）是一个评分器，用于计算满足所有必须（必须出现）子查询的文档的得分。
// 它将文档与每个子查询进行匹配，并将匹配的文档的得分进行累加。ReqScorer的得分是所有必须子查询的得分之和。
//
// OptScorer（可选匹配评分器）是一个评分器，用于计算满足任意可选（可选出现）子查询的文档的得分。
// 它将文档与每个可选子查询进行匹配，并将匹配的文档的得分进行累加。OptScorer的得分是所有可选子查询的得分之和。
//
// ReqOptSumScorer将ReqScorer和OptScorer的得分进行相加，得到最终的文档得分。
// 这意味着文档必须匹配所有必须子查询，并且可以匹配任意可选子查询。
//
// 使用ReqOptSumScorer可以实现布尔查询的组合逻辑，例如"must"（必须匹配）和"should"（可选匹配）的组合。
// 它允许您根据查询要求计算文档的相关性得分，并根据得分对文档进行排序和排名。
//
// 请注意，以上是对ReqOptSumScorer的基本解释，实际的实现和使用方式可能会有所不同，具体取决于Lucene版本和上下文环境。
type ReqOptSumScorer struct {
	*BaseScorer

	reqScorer search.Scorer
	optScorer search.Scorer

	reqApproximation types.DocIdSetIterator
	optApproximation types.DocIdSetIterator
	optTwoPhase      search.TwoPhaseIterator
	approximation    types.DocIdSetIterator
	twoPhase         search.TwoPhaseIterator

	maxScorePropagator *MaxScoreSumPropagator
	minScore           float64
	reqMaxScore        float64
	optIsRequired      bool
}

// NewReqOptSumScorer Construct a ReqOptScorer.
// reqScorer: The required scorer. This must match.
// optScorer: The optional scorer. This is used for scoring only.
// scoreMode: How the produced scorers will be consumed.
func NewReqOptSumScorer(reqScorer, optScorer search.Scorer, scoreMode search.ScoreMode) (*ReqOptSumScorer, error) {
	scorer := &ReqOptSumScorer{
		BaseScorer: NewScorer(reqScorer.GetWeight()),
		reqScorer:  reqScorer,
		optScorer:  optScorer,
	}

	if scoreMode == TOP_SCORES {
		sumPropagator, err := NewMaxScoreSumPropagator([]search.Scorer{reqScorer, optScorer})
		if err != nil {
			return nil, err
		}
		scorer.maxScorePropagator = sumPropagator
	}

	reqTwoPhase := reqScorer.TwoPhaseIterator()
	scorer.optTwoPhase = optScorer.TwoPhaseIterator()
	if reqTwoPhase == nil {
		scorer.reqApproximation = reqScorer.Iterator()
	} else {
		scorer.reqApproximation = reqTwoPhase.Approximation()
	}

	if scorer.optTwoPhase == nil {
		scorer.optApproximation = optScorer.Iterator()
	} else {
		scorer.optApproximation = scorer.optTwoPhase.Approximation()
	}

	if scoreMode != TOP_SCORES {
		scorer.approximation = scorer.reqApproximation
		scorer.reqMaxScore = math.Inf(0)
	} else {
		_, err := reqScorer.AdvanceShallow(0)
		if err != nil {
			return nil, err
		}
		_, err = optScorer.AdvanceShallow(0)
		if err != nil {
			return nil, err
		}
		scorer.reqMaxScore, err = reqScorer.GetMaxScore(types.NO_MORE_DOCS)
		if err != nil {
			return nil, err
		}
		scorer.approximation = &innerDocIdSetIterator{
			upTo:     -1,
			maxScore: 0,
			scorer:   scorer,
		}
	}

	if reqTwoPhase == nil && scorer.optTwoPhase == nil {
		scorer.twoPhase = nil
	} else {
		scorer.twoPhase = &innerTwoPhaseIterator{
			approximation: scorer.approximation,
			scorer:        scorer,
			reqTwoPhase:   reqTwoPhase,
		}
	}
	return scorer, nil
}

var _ types.DocIdSetIterator = &innerDocIdSetIterator{}

type innerDocIdSetIterator struct {
	upTo     int
	maxScore float64
	scorer   *ReqOptSumScorer
}

func (r *innerDocIdSetIterator) moveToNextBlock(target int) (err error) {
	r.upTo, err = r.scorer.AdvanceShallow(target)
	if err != nil {
		return err
	}
	reqMaxScoreBlock, err := r.scorer.reqScorer.GetMaxScore(r.upTo)
	r.maxScore, err = r.scorer.GetMaxScore(r.upTo)

	// Potentially move to a conjunction
	r.scorer.optIsRequired = reqMaxScoreBlock < r.scorer.minScore
	return nil
}

func (r *innerDocIdSetIterator) advanceImpacts(target int) (int, error) {
	if target > r.upTo {
		err := r.moveToNextBlock(target)
		if err != nil {
			return 0, err
		}
	}

	for {
		if r.maxScore >= r.scorer.minScore {
			return target, nil
		}

		if r.upTo == types.NO_MORE_DOCS {
			return types.NO_MORE_DOCS, nil
		}

		target = r.upTo + 1

		err := r.moveToNextBlock(target)
		if err != nil {
			return 0, err
		}
	}
}

func (r *innerDocIdSetIterator) DocID() int {
	return r.scorer.reqApproximation.DocID()
}

func (r *innerDocIdSetIterator) NextDoc() (int, error) {
	return r.advanceInternal(r.scorer.reqApproximation.DocID() + 1)
}

func (r *innerDocIdSetIterator) Advance(target int) (int, error) {
	return r.advanceInternal(target)
}

func (r *innerDocIdSetIterator) advanceInternal(target int) (int, error) {
	if target == types.NO_MORE_DOCS {
		if _, err := r.scorer.reqApproximation.Advance(target); err != nil {
			return 0, err
		}
		return types.NO_MORE_DOCS, nil
	}

	reqDoc := target

	var err error

OUTER:
	for {
		if r.scorer.minScore != 0 {
			reqDoc, err = r.advanceImpacts(reqDoc)
			if err != nil {
				return 0, err
			}
		}
		if r.scorer.reqApproximation.DocID() < reqDoc {
			reqDoc, err = r.scorer.reqApproximation.Advance(reqDoc)
			if err != nil {
				return 0, err
			}
		}
		if reqDoc == types.NO_MORE_DOCS || r.scorer.optIsRequired == false {
			return reqDoc, nil
		}

		upperBound := r.upTo
		if r.scorer.reqMaxScore < r.scorer.minScore {
			upperBound = types.NO_MORE_DOCS
		}
		if reqDoc > upperBound {
			continue
		}

		// Find the next common doc within the current block
		for {
			optDoc := r.scorer.optApproximation.DocID()
			if optDoc < reqDoc {
				optDoc, err = r.scorer.optApproximation.Advance(reqDoc)
				if err != nil {
					return 0, err
				}
			}
			if optDoc > upperBound {
				reqDoc = upperBound + 1
				continue OUTER
			}

			if optDoc != reqDoc {
				reqDoc, err = r.scorer.reqApproximation.Advance(optDoc)
				if err != nil {
					return 0, err
				}
				if reqDoc > upperBound {
					continue OUTER
				}
			}

			if reqDoc == types.NO_MORE_DOCS || optDoc == reqDoc {
				return reqDoc, nil
			}
		}

	}
}

func (r *innerDocIdSetIterator) SlowAdvance(target int) (int, error) {
	return types.SlowAdvance(r, target)
}

func (r *innerDocIdSetIterator) Cost() int64 {
	return r.scorer.reqApproximation.Cost()
}

var _ search.TwoPhaseIterator = &innerTwoPhaseIterator{}

type innerTwoPhaseIterator struct {
	approximation types.DocIdSetIterator
	scorer        *ReqOptSumScorer
	reqTwoPhase   search.TwoPhaseIterator
}

func (i *innerTwoPhaseIterator) Approximation() types.DocIdSetIterator {
	return i.approximation
}

func (i *innerTwoPhaseIterator) Matches() (bool, error) {
	matchValues, err := i.reqTwoPhase.Matches()
	if err != nil {
		return false, err
	}

	if i.reqTwoPhase != nil && matchValues == false {
		return false, nil
	}

	if i.scorer.optTwoPhase != nil {
		if i.scorer.optIsRequired {
			// The below condition is rare and can only happen if we transitioned to optIsRequired=true
			// after the opt approximation was advanced and before it was confirmed.
			if i.scorer.reqScorer.DocID() != i.scorer.optApproximation.DocID() {
				if i.scorer.optApproximation.DocID() < i.scorer.reqScorer.DocID() {
					if _, err := i.scorer.optApproximation.Advance(i.scorer.reqScorer.DocID()); err != nil {
						return false, err
					}
				}
				if i.scorer.reqScorer.DocID() != i.scorer.optApproximation.DocID() {
					return false, nil
				}
			}
			if ok, _ := i.scorer.optTwoPhase.Matches(); !ok {
				// Advance the iterator to make it clear it doesn't match the current doc id
				if _, err := i.scorer.optApproximation.NextDoc(); err != nil {
					return false, err
				}
				return false, nil
			}
		} else if match, _ := i.scorer.optTwoPhase.Matches(); i.scorer.optApproximation.DocID() == i.scorer.reqScorer.DocID() && match == false {
			// Advance the iterator to make it clear it doesn't match the current doc id
			if _, err := i.scorer.optApproximation.NextDoc(); err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func (i *innerTwoPhaseIterator) MatchCost() float64 {
	cost := 1.0
	if i.reqTwoPhase != nil {
		cost += i.reqTwoPhase.MatchCost()
	}
	if i.scorer.optTwoPhase != nil {
		cost += i.scorer.optTwoPhase.MatchCost()
	}
	return cost
}

func (r *ReqOptSumScorer) TwoPhaseIterator() search.TwoPhaseIterator {
	return r.twoPhase
}

func (r *ReqOptSumScorer) Score() (float64, error) {
	// TODO: sum into a double and cast to float if we ever send required clauses to BS1
	curDoc := r.reqScorer.DocID()
	score, err := r.reqScorer.Score()
	if err != nil {
		return 0, err
	}

	optScorerDoc := r.optApproximation.DocID()
	if optScorerDoc < curDoc {
		optScorerDoc, err = r.optApproximation.Advance(curDoc)
		if err != nil {
			return 0, err
		}
		if match, _ := r.optTwoPhase.Matches(); r.optTwoPhase != nil && optScorerDoc == curDoc && match == false {
			optScorerDoc, err = r.optApproximation.NextDoc()
			if err != nil {
				return 0, err
			}
		}
	}

	if optScorerDoc == curDoc {
		num, err := r.optScorer.Score()
		if err != nil {
			return 0, err
		}
		score += num
	}
	return score, nil
}

func (r *ReqOptSumScorer) DocID() int {
	return r.reqScorer.DocID()
}

func (r *ReqOptSumScorer) Iterator() types.DocIdSetIterator {
	if r.twoPhase == nil {
		return r.approximation
	} else {
		return AsDocIdSetIterator(r.twoPhase)
	}
}

func (r *ReqOptSumScorer) AdvanceShallow(target int) (int, error) {
	upTo, err := r.reqScorer.AdvanceShallow(target)
	if err != nil {
		return 0, err
	}
	if r.optScorer.DocID() <= target {
		shallow, err := r.optScorer.AdvanceShallow(target)
		if err != nil {
			return 0, err
		}
		upTo = min(upTo, shallow)
	} else if r.optScorer.DocID() != types.NO_MORE_DOCS {
		upTo = min(upTo, r.optScorer.DocID()-1)
	}
	return upTo, nil
}

func (r *ReqOptSumScorer) GetMaxScore(upTo int) (float64, error) {
	maxScore, err := r.reqScorer.GetMaxScore(upTo)
	if err != nil {
		return 0, err
	}
	if r.optScorer.DocID() <= upTo {
		num, err := r.optScorer.GetMaxScore(upTo)
		if err != nil {
			return 0, err
		}
		maxScore += num
	}
	return maxScore, nil
}

func (r *ReqOptSumScorer) SetMinCompetitiveScore(minScore float64) error {
	r.minScore = minScore
	// Potentially move to a conjunction
	if r.reqMaxScore < minScore {
		r.optIsRequired = true
	}
	return r.maxScorePropagator.SetMinCompetitiveScore(minScore)
}
