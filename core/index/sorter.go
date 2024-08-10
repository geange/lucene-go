package index

import "github.com/geange/lucene-go/core/interface/index"

const (
	BINARY_SORT_THRESHOLD    = 20
	INSERTION_SORT_THRESHOLD = 16
)

type SorterDefault struct {
	pivotIndex int
	fnCompare  func(i, j int) int
	fnSwap     func(i, j int)
}

func SortByComparator(maxDoc int, comparator index.DocComparator) index.DocMap {
	// TODO: fix it
	panic("")
	/*

		// sort doc IDs
		docs := make([]int, maxDoc)
		for i := range docs {
			docs[i] = i
		}

		sorter := NewDocValueSorter(docs, comparator)
		sort.Sort(sorter)

		newToOldBuilder := packed.NewPackedLongValuesBuilder()
		for i := 0; i < maxDoc; i++ {
			newToOldBuilder.Add(int64(docs[i]))
		}

		newToOld := newToOldBuilder.Build()

		// invert the docs mapping:
		for i := 0; i < maxDoc; i++ {
			docs[newToOld.Get(int64(i))] = i
		} // docs is now the oldToNew mapping

		oldToNewBuilder := packed.NewPackedLongValuesBuilder()
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

	*/

}

var _ index.DocComparator = &EmptyDocComparator{}

type EmptyDocComparator struct {
	FnCompare func(docID1, docID2 int) int
}

func (e *EmptyDocComparator) Compare(docID1, docID2 int) int {
	return e.FnCompare(docID1, docID2)
}

func SortByComparators(maxDoc int, comparators []index.DocComparator) (index.DocMap, error) {
	// TODO: fix it
	/*
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

	*/
	panic("")
}
