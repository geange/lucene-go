package search

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util/structure"
	"math"
)

// CreateFieldValueHitQueue
// Creates a hit queue sorted by the given list of fields.
// NOTE: The instances returned by this method pre-allocate a full array of length numHits.
// Params: 	fields – SortField array we are sorting by in priority order (highest priority first);
//
//			 cannot be null or empty
//	size – The number of hits to retain. Must be greater than zero.
func CreateFieldValueHitQueue(fields []index.SortField, size int) FieldValueHitQueue[*Entry] {
	if len(fields) == 1 {
		return NewOneComparatorFieldValueHitQueue(fields, size)
	}
	return NewMultiComparatorsFieldValueHitQueue(fields, size)
}

// FieldValueHitQueue
// Expert: A hit queue for sorting by hits by terms in more than one field.
// Since: 2.9
// See Also: IndexSearcher.search(Query, int, Sort)
// lucene.experimental
type FieldValueHitQueue[T index.ScoreDoc] interface {
	Add(element T) T
	Top() T
	Pop() (T, error)
	UpdateTop() T
	UpdateTopByNewTop(newTop T) T
	Size() int
	Clear()
	Remove(element T) bool
	Iterator() structure.Iterator[T]
	GetReverseMul() []int
	GetComparators(ctx index.LeafReaderContext) ([]index.LeafFieldComparator, error)
	GetComparatorsList() []index.FieldComparator
}

type FieldValueHitQueueDefault[T any] struct {
	*structure.PriorityQueue[T]

	//Stores the sort criteria being used.
	fields      []index.SortField
	comparators []index.FieldComparator
	reverseMul  []int
}

// newFieldValueHitQueue
// prevent instantiation and extension.
func newFieldValueHitQueue[T any](fields []index.SortField, size int, lessThan func(a, b T) bool) *FieldValueHitQueueDefault[T] {
	pq := structure.NewPriorityQueue(size, lessThan)

	// When we get here, fields.length is guaranteed to be > 0, therefore no
	// need to check it again.

	// All these are required by this class's API - need to return arrays.
	// Therefore even in the case of a single comparator, create an array
	// anyway.
	queue := &FieldValueHitQueueDefault[T]{
		PriorityQueue: pq,
	}
	queue.fields = fields
	numComparators := len(fields)
	queue.comparators = make([]index.FieldComparator, numComparators)
	queue.reverseMul = make([]int, numComparators)
	for i := 0; i < numComparators; i++ {
		field := fields[i]
		queue.reverseMul[i] = 1
		if field.GetReverse() {
			queue.reverseMul[i] = -1
		}
		queue.comparators[i] = field.GetComparator(size, i)
	}
	return queue
}

func (f *FieldValueHitQueueDefault[T]) GetComparators(ctx index.LeafReaderContext) ([]index.LeafFieldComparator, error) {
	comparators := make([]index.LeafFieldComparator, 0)
	for _, comparator := range f.comparators {
		leafComparator, err := comparator.GetLeafComparator(ctx)
		if err != nil {
			return nil, err
		}
		comparators = append(comparators, leafComparator)
	}
	return comparators, nil
}

func (f *FieldValueHitQueueDefault[T]) GetReverseMul() []int {
	return f.reverseMul
}

func (f *FieldValueHitQueueDefault[T]) GetComparatorsList() []index.FieldComparator {
	return f.comparators
}

type Entry struct {
	*baseScoreDoc

	slot int
}

func NewEntry(slot, doc int) *Entry {
	return &Entry{
		baseScoreDoc: newScoreDoc(doc, math.NaN()),
		slot:         slot,
	}
}

type OneComparatorFieldValueHitQueue struct {
	oneReverseMul int
	oneComparator index.FieldComparator
	*FieldValueHitQueueDefault[*Entry]
}

func NewOneComparatorFieldValueHitQueue(fields []index.SortField, size int) *OneComparatorFieldValueHitQueue {
	queue := &OneComparatorFieldValueHitQueue{}
	queue.FieldValueHitQueueDefault = newFieldValueHitQueue(fields, size, queue.Less)
	queue.oneComparator = queue.comparators[0]
	queue.oneReverseMul = queue.reverseMul[0]
	return queue
}

func (o *OneComparatorFieldValueHitQueue) Less(hitA, hitB *Entry) bool {
	c := o.oneReverseMul * o.oneComparator.Compare(hitA.slot, hitB.slot)
	if c != 0 {
		return c > 0
	}

	// avoid random sort order that could lead to duplicates (bug #31241):
	return hitA.GetDoc() > hitB.GetDoc()
}

type MultiComparatorsFieldValueHitQueue struct {
	*FieldValueHitQueueDefault[*Entry]
}

func NewMultiComparatorsFieldValueHitQueue(fields []index.SortField, size int) *MultiComparatorsFieldValueHitQueue {
	queue := &MultiComparatorsFieldValueHitQueue{}
	queue.FieldValueHitQueueDefault = newFieldValueHitQueue(fields, size, queue.Less)
	return queue
}

func (m *MultiComparatorsFieldValueHitQueue) Less(hitA, hitB *Entry) bool {
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
