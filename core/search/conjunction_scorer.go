package search

import (
	"errors"
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/index"
	"math"
	"sort"
)

var _ Scorer = &ConjunctionScorer{}

// ConjunctionScorer
// Create a new ConjunctionScorer, note that scorers must be a subset of required.
type ConjunctionScorer struct {
	*ScorerDefault

	disi     index.DocIdSetIterator
	scorers  []Scorer
	required []Scorer
}

func NewConjunctionScorer(weight Weight, scorers []Scorer, required []Scorer) (*ConjunctionScorer, error) {
	disi, err := intersectScorers(scorers)
	if err != nil {
		return nil, err
	}
	return &ConjunctionScorer{
		ScorerDefault: NewScorer(weight),
		disi:          disi,
		scorers:       scorers,
		required:      required,
	}, nil
}

func (c *ConjunctionScorer) Score() (float64, error) {
	sum := 0.0
	for _, scorer := range c.scorers {
		v, err := scorer.Score()
		if err != nil {
			return 0, err
		}
		sum += v
	}
	return sum, nil
}

func (c *ConjunctionScorer) DocID() int {
	docID := c.disi.DocID()
	if err != nil {
		return 0
	}
	return docID
}

func (c *ConjunctionScorer) TwoPhaseIterator() TwoPhaseIterator {
	return UnwrapIterator(c.disi)
}

func (c *ConjunctionScorer) Iterator() index.DocIdSetIterator {
	return c.disi
}

func (c *ConjunctionScorer) GetMaxScore(upTo int) (float64, error) {
	switch len(c.scorers) {
	case 0:
		return 0, nil
	case 1:
		return c.scorers[0].GetMaxScore(upTo)
	default:
		return math.Inf(-1), nil
	}
}

func intersectScorers(scorers []Scorer) (index.DocIdSetIterator, error) {
	if len(scorers) < 2 {
		return nil, errors.New("cannot make a ConjunctionDISI of less than 2 iterators")
	}

	allIterators := make([]index.DocIdSetIterator, 0)
	twoPhaseIterators := make([]TwoPhaseIterator, 0)

	for _, scorer := range scorers {
		allIterators, twoPhaseIterators = addScorer(scorer, allIterators, twoPhaseIterators)
	}
	return createConjunction(allIterators, twoPhaseIterators)
}

// Adds the scorer, possibly splitting up into two phases or collapsing if it is another conjunction
func addScorer(scorer Scorer, allIterators []index.DocIdSetIterator,
	twoPhaseIterators []TwoPhaseIterator) ([]index.DocIdSetIterator, []TwoPhaseIterator) {
	twoPhaseIter := scorer.TwoPhaseIterator()
	if twoPhaseIter != nil {
		allIterators, twoPhaseIterators = addTwoPhaseIterator(twoPhaseIter, allIterators, twoPhaseIterators)
	} else {
		// no approximation support, use the iterator as-is
		// TODO
	}
	return allIterators, twoPhaseIterators
}

func addTwoPhaseIterator(twoPhaseIter TwoPhaseIterator, allIterators []index.DocIdSetIterator,
	twoPhaseIterators []TwoPhaseIterator) ([]index.DocIdSetIterator, []TwoPhaseIterator) {
	allIterators, twoPhaseIterators = addIterator(twoPhaseIter.Approximation(), allIterators, twoPhaseIterators)
	if v, ok := twoPhaseIter.(*ConjunctionTwoPhaseIterator); ok {
		// Check for exactly this class for collapsing
		twoPhaseIterators = append(twoPhaseIterators, v.twoPhaseIterators...)
	} else {
		twoPhaseIterators = append(twoPhaseIterators, twoPhaseIter)
	}
	return allIterators, twoPhaseIterators
}

func createConjunction(allIterators []index.DocIdSetIterator,
	twoPhaseIterators []TwoPhaseIterator) (index.DocIdSetIterator, error) {

	// check that all sub-iterators are on the same doc ID
	curDoc := 0
	if len(allIterators) > 0 {
		curDoc = allIterators[0].DocID()
	} else {
		twoPhaseIterators[0].Approximation().DocID()
	}

	iteratorsOnTheSameDoc := true
	for _, iterator := range allIterators {
		docID := iterator.DocID()
		if err != nil {
			return nil, err
		}
		if docID != curDoc {
			iteratorsOnTheSameDoc = false
			break
		}
	}

	if iteratorsOnTheSameDoc {
		match := true
		for _, iterator := range twoPhaseIterators {
			docID := iterator.Approximation().DocID()
			if err != nil {
				return nil, err
			}
			if docID != curDoc {
				match = false
				break
			}
		}

		iteratorsOnTheSameDoc = match
	}

	if !iteratorsOnTheSameDoc {
		return nil, errors.New("sub-iterators of ConjunctionDISI are not on the same document")
	}

	minCost := int64(math.MaxInt64)
	for _, iterator := range allIterators {
		cost := iterator.Cost()
		if cost < minCost {
			minCost = cost
		}
	}

	bitSetIterators := make([]*index.BitSetIterator, 0)
	iterators := make([]index.DocIdSetIterator, 0)

	for _, it := range allIterators {
		if v, ok := it.(*index.BitSetIterator); ok {
			if it.Cost() > minCost {
				bitSetIterators = append(bitSetIterators, v)
				continue
			}
		}
		iterators = append(iterators, it)
	}

	var disi index.DocIdSetIterator
	if len(iterators) == 1 {
		disi = iterators[0]
	} else {
		disi = newConjunctionDISI(iterators)
	}

	if len(bitSetIterators) > 0 {
		disi = newBitSetConjunctionDISI(disi, bitSetIterators)
	}

	if len(twoPhaseIterators) > 0 {
		disi = AsDocIdSetIterator(newConjunctionTwoPhaseIterator(disi, twoPhaseIterators))
	}
	return disi, nil
}

func addIterator(disi index.DocIdSetIterator, allIterators []index.DocIdSetIterator,
	twoPhaseIterators []TwoPhaseIterator) ([]index.DocIdSetIterator, []TwoPhaseIterator) {

	twoPhase := UnwrapIterator(disi)

	if twoPhase != nil {
		allIterators, twoPhaseIterators = addTwoPhaseIterator(twoPhase, allIterators, twoPhaseIterators)
	} else if conjunction, ok := disi.(*ConjunctionDISI); ok {
		// Check for exactly this class for collapsing
		// subconjuctions have already split themselves into two phase iterators and others, so we can take those
		// iterators as they are and move them up to this conjunction
		allIterators = append(allIterators, conjunction.lead1, conjunction.lead2)
		allIterators = append(allIterators, conjunction.others...)
	} else if conjunction, ok := disi.(*BitSetConjunctionDISI); ok {
		allIterators = append(allIterators, conjunction.lead)
		for _, iterator := range conjunction.bitSetIterators {
			allIterators = append(allIterators, iterator)
		}
	} else {
		allIterators = append(allIterators, disi)
	}
	return allIterators, twoPhaseIterators
}

var _ TwoPhaseIterator = &ConjunctionTwoPhaseIterator{}

type ConjunctionTwoPhaseIterator struct {
	approximation     index.DocIdSetIterator
	twoPhaseIterators []TwoPhaseIterator
	matchCost         float64
}

var _ sort.Interface = TimSortTwoPhase{}

type TimSortTwoPhase []TwoPhaseIterator

func (t TimSortTwoPhase) Len() int {
	return len(t)
}

func (t TimSortTwoPhase) Less(i, j int) bool {
	return t[i].MatchCost() < t[j].MatchCost()
}

func (t TimSortTwoPhase) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func newConjunctionTwoPhaseIterator(approximation index.DocIdSetIterator,
	twoPhaseIterators []TwoPhaseIterator) *ConjunctionTwoPhaseIterator {

	it := &ConjunctionTwoPhaseIterator{approximation: approximation}

	sort.Sort(TimSortTwoPhase(twoPhaseIterators))

	it.twoPhaseIterators = twoPhaseIterators

	// Compute the matchCost as the total matchCost of the sub iterators.
	// TODO: This could be too high because the matching is done cheapest first: give the lower matchCosts a higher weight.
	totalMatchCost := 0.0
	for _, tpi := range twoPhaseIterators {
		totalMatchCost += tpi.MatchCost()
	}
	it.matchCost = totalMatchCost
	return it
}

func (c *ConjunctionTwoPhaseIterator) Approximation() index.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}

func (c *ConjunctionTwoPhaseIterator) Matches() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConjunctionTwoPhaseIterator) MatchCost() float64 {
	return c.matchCost
}

var _ index.DocIdSetIterator = &BitSetConjunctionDISI{}

type BitSetConjunctionDISI struct {
	lead            index.DocIdSetIterator
	bitSetIterators []*index.BitSetIterator
	bitSets         []*bitset.BitSet
	minLength       int
}

var _ sort.Interface = TimSortBitSet{}

type TimSortBitSet []*index.BitSetIterator

func (t TimSortBitSet) Len() int {
	return len(t)
}

func (t TimSortBitSet) Less(i, j int) bool {
	return t[i].Cost() < t[j].Cost()
}

func (t TimSortBitSet) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func newBitSetConjunctionDISI(lead index.DocIdSetIterator, bitSetIterators []*index.BitSetIterator) *BitSetConjunctionDISI {
	disi := &BitSetConjunctionDISI{
		lead:            lead,
		bitSetIterators: bitSetIterators,
	}
	sort.Sort(TimSortBitSet(bitSetIterators))
	disi.bitSets = make([]*bitset.BitSet, len(bitSetIterators))
	minLen := math.MaxInt64
	for i, iterator := range disi.bitSetIterators {
		bitSet := iterator.GetBitSet()
		disi.bitSets[i] = bitSet
		minLen = min(minLen, int(bitSet.Count()))
	}
	disi.minLength = minLen
	return disi
}

func (b *BitSetConjunctionDISI) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (b *BitSetConjunctionDISI) NextDoc() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BitSetConjunctionDISI) Advance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BitSetConjunctionDISI) SlowAdvance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BitSetConjunctionDISI) Cost() int64 {
	//TODO implement me
	panic("implement me")
}
