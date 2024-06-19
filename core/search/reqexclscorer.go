package search

import (
	"github.com/geange/lucene-go/core/interface/search"
	"github.com/geange/lucene-go/core/types"
)

var _ search.Scorer = &ReqExclScorer{}

// ReqExclScorer
// A Scorer for queries with a required subscorer and an excluding (prohibited) sub Scorer.
type ReqExclScorer struct {
	*BaseScorer

	reqScorer search.Scorer

	// approximations of the scorers, or the scorers themselves if they don't support approximations
	reqApproximation types.DocIdSetIterator

	exclApproximation types.DocIdSetIterator

	// two-phase views of the scorers, or null if they do not support approximations
	reqTwoPhaseIterator search.TwoPhaseIterator

	exclTwoPhaseIterator search.TwoPhaseIterator
}

func NewReqExclScorer(reqScorer, exclScorer search.Scorer) *ReqExclScorer {
	scorer := &ReqExclScorer{
		BaseScorer:          NewScorer(reqScorer.GetWeight()),
		reqScorer:           reqScorer,
		reqTwoPhaseIterator: reqScorer.TwoPhaseIterator(),
	}

	if scorer.reqTwoPhaseIterator == nil {
		scorer.reqApproximation = reqScorer.Iterator()
	} else {
		scorer.reqApproximation = scorer.reqTwoPhaseIterator.Approximation()
	}
	scorer.exclTwoPhaseIterator = exclScorer.TwoPhaseIterator()
	if scorer.exclTwoPhaseIterator == nil {
		scorer.exclApproximation = exclScorer.Iterator()
	} else {
		scorer.exclApproximation = scorer.exclTwoPhaseIterator.Approximation()
	}
	return scorer
}

func (r *ReqExclScorer) Score() (float64, error) {
	return r.reqScorer.Score()
}

func (r *ReqExclScorer) DocID() int {
	return r.reqApproximation.DocID()
}

func (r *ReqExclScorer) Iterator() types.DocIdSetIterator {
	return AsDocIdSetIterator(r.TwoPhaseIterator())
}

func (r *ReqExclScorer) GetMaxScore(upTo int) (float64, error) {
	return r.reqScorer.GetMaxScore(upTo)
}

const (
	ADVANCE_COST = 10
)

func matchCost(reqApproximation types.DocIdSetIterator, reqTwoPhaseIterator search.TwoPhaseIterator,
	exclApproximation types.DocIdSetIterator, exclTwoPhaseIterator search.TwoPhaseIterator) float64 {

	matchCostVar := float64(2) // we perform 2 comparisons to advance exclApproximation
	if reqTwoPhaseIterator != nil {
		// this two-phase iterator must always be matched
		matchCostVar += reqTwoPhaseIterator.MatchCost()
	}

	// match cost of the prohibited clause: we need to advance the approximation
	// and match the two-phased iterator
	exclMatchCost := float64(ADVANCE_COST)
	if exclTwoPhaseIterator != nil {
		exclMatchCost += exclTwoPhaseIterator.MatchCost()
	}

	// upper value for the ratio of documents that reqApproximation matches that
	// exclApproximation also matches
	ratio := float64(0)
	if reqApproximation.Cost() <= 0 {
		ratio = 1
	} else if exclApproximation.Cost() <= 0 {
		ratio = 0
	} else {
		ratio = float64(min(reqApproximation.Cost(), exclApproximation.Cost())) / float64(reqApproximation.Cost())
	}
	matchCostVar += ratio * exclMatchCost

	return matchCostVar
}

func (r *ReqExclScorer) TwoPhaseIterator() search.TwoPhaseIterator {
	cost := matchCost(r.reqApproximation, r.reqTwoPhaseIterator, r.exclApproximation, r.exclTwoPhaseIterator)

	if r.reqTwoPhaseIterator == nil ||
		(r.exclTwoPhaseIterator != nil && r.reqTwoPhaseIterator.MatchCost() <= r.exclTwoPhaseIterator.MatchCost()) {
		// reqTwoPhaseIterator is LESS costly than exclTwoPhaseIterator, check it first
		return &twoPhaseIterator1{
			reqApproximation:     r.reqApproximation,
			reqTwoPhaseIterator:  r.reqTwoPhaseIterator,
			exclApproximation:    r.exclApproximation,
			exclTwoPhaseIterator: r.exclTwoPhaseIterator,
			matchCost:            cost,
		}
	} else {
		// reqTwoPhaseIterator is MORE costly than exclTwoPhaseIterator, check it last
		return &twoPhaseIterator2{
			reqApproximation:     r.reqApproximation,
			reqTwoPhaseIterator:  r.reqTwoPhaseIterator,
			exclApproximation:    r.exclApproximation,
			exclTwoPhaseIterator: r.exclTwoPhaseIterator,
			matchCost:            cost,
		}
	}
}

var _ search.TwoPhaseIterator = &twoPhaseIterator1{}

type twoPhaseIterator1 struct {
	reqApproximation     types.DocIdSetIterator
	reqTwoPhaseIterator  search.TwoPhaseIterator
	exclApproximation    types.DocIdSetIterator
	exclTwoPhaseIterator search.TwoPhaseIterator
	matchCost            float64
}

func (t *twoPhaseIterator1) Approximation() types.DocIdSetIterator {
	return t.reqApproximation
}

func (t *twoPhaseIterator1) Matches() (bool, error) {
	var err error

	doc := t.reqApproximation.DocID()
	// check if the doc is not excluded
	exclDoc := t.exclApproximation.DocID()
	if exclDoc < doc {
		exclDoc, err = t.exclApproximation.Advance(doc)
		if err != nil {
			return false, err
		}
	}

	if exclDoc < doc {
		return matchesOrNull(t.reqTwoPhaseIterator)
	}
	m1, err := matchesOrNull(t.reqTwoPhaseIterator)
	if err != nil {
		return false, err
	}
	m2, err := matchesOrNull(t.exclTwoPhaseIterator)
	if err != nil {
		return false, err
	}
	return m1 && !m2, nil
}

func (t *twoPhaseIterator1) MatchCost() float64 {
	return t.matchCost
}

// Confirms whether or not the given TwoPhaseIterator matches on the current document.
func matchesOrNull(it search.TwoPhaseIterator) (bool, error) {
	if it == nil {
		return false, nil
	}

	ok, err := it.Matches()
	if err != nil {
		return false, err
	}
	return ok, nil
}

var _ search.TwoPhaseIterator = &twoPhaseIterator2{}

type twoPhaseIterator2 struct {
	reqApproximation     types.DocIdSetIterator
	reqTwoPhaseIterator  search.TwoPhaseIterator
	exclApproximation    types.DocIdSetIterator
	exclTwoPhaseIterator search.TwoPhaseIterator
	matchCost            float64
}

func (t *twoPhaseIterator2) Approximation() types.DocIdSetIterator {
	return t.reqApproximation
}

func (t *twoPhaseIterator2) Matches() (bool, error) {
	var err error
	doc := t.reqApproximation.DocID()
	// check if the doc is not excluded
	exclDoc := t.exclApproximation.DocID()
	if exclDoc < doc {
		exclDoc, err = t.exclApproximation.Advance(doc)
		if err != nil {
			return false, err
		}
	}
	if exclDoc != doc {
		return matchesOrNull(t.reqTwoPhaseIterator)
	}
	m1, err := matchesOrNull(t.exclTwoPhaseIterator)
	if err != nil {
		return false, err
	}
	m2, err := matchesOrNull(t.reqTwoPhaseIterator)
	if err != nil {
		return false, err
	}
	return !m1 && m2, nil
}

func (t *twoPhaseIterator2) MatchCost() float64 {
	return t.matchCost
}
