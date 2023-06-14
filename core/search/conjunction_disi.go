package search

import (
	"github.com/geange/lucene-go/core/index"
	"sort"
)

var _ index.DocIdSetIterator = &ConjunctionDISI{}

// ConjunctionDISI
// A conjunction of DocIdSetIterators. Requires that all of its sub-iterators must be on the same
// document all the time. This iterates over the doc ids that are present in each given DocIdSetIterator.
// Public only for use in org.apache.lucene.search.spans.
// lucene.internal
type ConjunctionDISI struct {
	lead1  index.DocIdSetIterator
	lead2  index.DocIdSetIterator
	others []index.DocIdSetIterator
}

func newConjunctionDISI(iterators []index.DocIdSetIterator) *ConjunctionDISI {
	// Sort the array the first time to allow the least frequent DocsEnum to
	// lead the matching.
	sort.Sort(TimSort(iterators))
	return &ConjunctionDISI{
		lead1:  iterators[0],
		lead2:  iterators[1],
		others: iterators[2:],
	}
}

var _ sort.Interface = TimSort{}

type TimSort []index.DocIdSetIterator

func (t TimSort) Len() int {
	return len(t)
}

func (t TimSort) Less(i, j int) bool {
	return t[i].Cost() < t[j].Cost()
}

func (t TimSort) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (c *ConjunctionDISI) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (c *ConjunctionDISI) NextDoc() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConjunctionDISI) Advance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConjunctionDISI) SlowAdvance(target int) (int, error) {
	return index.SlowAdvance(c, target)
}

func (c *ConjunctionDISI) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

// IntersectIterators Create a conjunction over the provided Scorers. Note that the returned DocIdSetIterator might leverage two-phase iteration in which case it is possible to retrieve the TwoPhaseIterator using TwoPhaseIterator.unwrap.
func IntersectIterators(iterators []index.DocIdSetIterator) index.DocIdSetIterator {
	panic("")
}

func (c *ConjunctionDISI) doNext(doc int) (int, error) {
	panic("")
}
