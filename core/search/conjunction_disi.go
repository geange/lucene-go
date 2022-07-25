package search

import "github.com/geange/lucene-go/core/index"

var _ index.DocIdSetIterator = &ConjunctionDISI{}

type ConjunctionDISI struct {
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
	//TODO implement me
	panic("implement me")
}

func (c *ConjunctionDISI) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

// IntersectIterators Create a conjunction over the provided Scorers. Note that the returned DocIdSetIterator might leverage two-phase iteration in which case it is possible to retrieve the TwoPhaseIterator using TwoPhaseIterator.unwrap.
func IntersectIterators(iterators []index.DocIdSetIterator) index.DocIdSetIterator {
	panic("")
}
