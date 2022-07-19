package memory

import "github.com/geange/lucene-go/core/index"

var _ index.SortedNumericDocValues = &innerSortedNumericDocValues{}

type innerSortedNumericDocValues struct {
	it     *memoryDocValuesIterator
	ord    int
	values []int
	count  int
}

func newInnerSortedNumericDocValues(values []int, count int) *innerSortedNumericDocValues {
	return &innerSortedNumericDocValues{
		it:     newMemoryDocValuesIterator(),
		ord:    0,
		values: values,
		count:  count,
	}
}

func (i *innerSortedNumericDocValues) DocID() int {
	return i.it.docId()
}

func (i *innerSortedNumericDocValues) NextDoc() (int, error) {
	return i.it.nextDoc(), nil
}

func (i *innerSortedNumericDocValues) Advance(target int) (int, error) {
	return i.it.advance(target), nil
}

func (i *innerSortedNumericDocValues) SlowAdvance(target int) (int, error) {
	doc := 0
	var err error
	for doc < target {
		doc, err = i.NextDoc()
		if err != nil {
			return 0, nil
		}
	}
	return doc, nil
}

func (i *innerSortedNumericDocValues) Cost() int64 {
	return 1
}

func (i *innerSortedNumericDocValues) AdvanceExact(target int) (bool, error) {
	advance, err := i.Advance(target)
	if err != nil {
		return false, err
	}
	return advance == target, nil
}

func (i *innerSortedNumericDocValues) NextValue() (int64, error) {
	ord := i.ord
	i.ord++
	return int64(i.values[ord]), nil
}

func (i *innerSortedNumericDocValues) DocValueCount() int {
	return i.count
}
