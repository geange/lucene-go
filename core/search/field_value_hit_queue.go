package search

import (
	"github.com/emirpasic/gods/queues/priorityqueue"
	"github.com/geange/lucene-go/core/index"
)

type FieldValueHitQueue[T ScoreDoc] interface {
	Less(a, b T) bool
}

type FieldValueHitQueueDefault struct {
	//Stores the sort criteria being used.
	fields      []index.SortField
	comparators []index.FieldComparator
	reverseMul  []int
}

// prevent instantiation and extension.
func NewFieldValueHitQueue(fields []index.SortField, size int) *FieldValueHitQueueDefault {

}

type Entry struct {
	*ScoreDocDefault
	slot int
}

var _ FieldValueHitQueue[*Entry] = &OneComparatorFieldValueHitQueue{}

type OneComparatorFieldValueHitQueue struct {
	oneReverseMul int
	oneComparator index.FieldComparator
	queue         *priorityqueue.Queue
}

func (o *OneComparatorFieldValueHitQueue) Less(hitA, hitB *Entry) bool {
	c := o.oneReverseMul * o.oneComparator.Compare(hitA.slot, hitB.slot)
	if c != 0 {
		return c > 0
	}

	// avoid random sort order that could lead to duplicates (bug #31241):
	return hitA.GetDoc() > hitB.GetDoc()
}

var _ FieldValueHitQueue[*Entry] = &MultiComparatorsFieldValueHitQueue{}

type MultiComparatorsFieldValueHitQueue struct {
	*FieldValueHitQueueDefault
	queue *priorityqueue.Queue
}

func (m MultiComparatorsFieldValueHitQueue) Less(hitA, hitB *Entry) bool {
	numComparators := len(m.comparators)
	for i := 0; i < numComparators; i++ {
		c := m.reverseMul[i] * m.comparators[i].Compare(hitA.slot, hitB.slot)
		if c != 0 {
			// Short circuit
			return c > 0
		}
	}

	// avoid random sort order that could lead to duplicates (bug #31241):
	return hitA.doc > hitB.doc
}
