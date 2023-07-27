package index

import (
	"github.com/geange/lucene-go/core/util/packed"
	"sort"
)

const (
	BINARY_SORT_THRESHOLD    = 20
	INSERTION_SORT_THRESHOLD = 16
)

// Sorter Base class for sorting algorithms implementations.
// lucene.internal
type Sorter interface {
	// Compare entries found in slots i and j. The contract for the returned item is the same as cmp.CompareFn(Object, Object).
	Compare(i, j int) int

	Swap(i, j int) int
}

type SorterDefault struct {
	pivotIndex int
	fnCompare  func(i, j int) int
	fnSwap     func(i, j int)
}

// DocMap A permutation of doc IDs. For every document ID between 0 and Reader.maxDoc(),
// oldToNew(newToOld(docID)) must return docID.
type DocMap struct {
	// Given a doc ID from the original index, return its ordinal in the sorted index.
	OldToNew func(docID int) int

	// Given the ordinal of a doc ID, return its doc ID in the original index.
	NewToOld func(docID int) int

	// Return the number of documents in this map.
	// This must be equal to the number of documents of the LeafReader which is sorted.
	Size func() int
}

func SortByComparator(maxDoc int, comparator DocComparator) *DocMap {

	// sort doc IDs
	docs := make([]int, maxDoc)
	for i := range docs {
		docs[i] = i
	}

	sorter := NewDocValueSorter(docs, comparator)
	sort.Sort(sorter)

	newToOldBuilder := packed.NewPackedLongValuesBuilderV1()
	for i := 0; i < maxDoc; i++ {
		newToOldBuilder.Add(int64(docs[i]))
	}

	newToOld := newToOldBuilder.Build()

	// invert the docs mapping:
	for i := 0; i < maxDoc; i++ {
		docs[newToOld.Get(int64(i))] = i
	} // docs is now the oldToNew mapping

	oldToNewBuilder := packed.NewPackedLongValuesBuilderV1()
	for i := 0; i < maxDoc; i++ {
		oldToNewBuilder.Add(int64(docs[i]))
	}
	oldToNew := oldToNewBuilder.Build()

	return &DocMap{
		OldToNew: func(docID int) int {
			return int(oldToNew.Get(int64(docID)))
		},
		NewToOld: func(docID int) int {
			return int(newToOld.Get(int64(docID)))
		},
		Size: func() int {
			return maxDoc
		},
	}
}

var _ DocComparator = &EmptyDocComparator{}

type EmptyDocComparator struct {
	FnCompare func(docID1, docID2 int) int
}

func (e *EmptyDocComparator) Compare(docID1, docID2 int) int {
	return e.FnCompare(docID1, docID2)
}

func SortByComparators(maxDoc int, comparators []DocComparator) (*DocMap, error) {
	return SortByComparator(maxDoc, &EmptyDocComparator{
		FnCompare: func(docID1, docID2 int) int {
			for _, comparator := range comparators {
				if cmp := comparator.Compare(docID1, docID2); cmp != 0 {
					return cmp
				}
			}
			if docID1 < docID2 {
				return 1
			} else if docID1 == docID2 {
				return 0
			} else {
				return -1
			}
		}}), nil
}
