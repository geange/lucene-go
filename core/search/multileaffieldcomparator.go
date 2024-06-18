package search

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.LeafFieldComparator = &MultiLeafFieldComparator{}

type MultiLeafFieldComparator struct {
	comparators []index.LeafFieldComparator
	reverseMul  []int

	// we extract the first comparator to avoid array access in the common case
	// that the first comparator compares worse than the bottom entry in the queue
	firstComparator index.LeafFieldComparator
	firstReverseMul int
}

func NewMultiLeafFieldComparator(comparators []index.LeafFieldComparator, reverseMul []int) *MultiLeafFieldComparator {
	panic("")
}

func (m *MultiLeafFieldComparator) SetBottom(slot int) error {
	//TODO implement me
	panic("implement me")
}

func (m *MultiLeafFieldComparator) CompareBottom(doc int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MultiLeafFieldComparator) CompareTop(doc int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MultiLeafFieldComparator) Copy(slot, doc int) error {
	//TODO implement me
	panic("implement me")
}

func (m *MultiLeafFieldComparator) SetScorer(scorer index.Scorable) error {
	//TODO implement me
	panic("implement me")
}

func (m *MultiLeafFieldComparator) CompetitiveIterator() (types.DocIdSetIterator, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MultiLeafFieldComparator) SetHitsThresholdReached() error {
	//TODO implement me
	panic("implement me")
}
