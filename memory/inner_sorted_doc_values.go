package memory

import (
	"github.com/geange/lucene-go/core/index"
)

var _ index.SortedDocValues = &innerSortedDocValues{}

type innerSortedDocValues struct {
	*index.SortedDocValuesImp

	value []byte
	it    *memoryDocValuesIterator
}

func newInnerSortedDocValues(value []byte) *innerSortedDocValues {
	values := &innerSortedDocValues{
		SortedDocValuesImp: nil,
		value:              value,
		it:                 newMemoryDocValuesIterator(),
	}
	values.SortedDocValuesImp = index.NewSortedDocValuesImp(values)
	return values
}

func (i *innerSortedDocValues) DocID() int {
	return i.it.docId()
}

func (i *innerSortedDocValues) NextDoc() (int, error) {
	return i.it.nextDoc(), nil
}

func (i *innerSortedDocValues) Advance(target int) (int, error) {
	return i.it.advance(target), nil
}

func (i *innerSortedDocValues) SlowAdvance(target int) (int, error) {
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

func (i *innerSortedDocValues) Cost() int64 {
	return 1
}

func (i *innerSortedDocValues) AdvanceExact(target int) (bool, error) {
	advance, err := i.Advance(target)
	if err != nil {
		return false, err
	}
	return advance == target, nil
}

func (i *innerSortedDocValues) OrdValue() (int, error) {
	return 0, nil
}

func (i *innerSortedDocValues) LookupOrd(ord int) ([]byte, error) {
	return i.value, nil
}

func (i *innerSortedDocValues) GetValueCount() int {
	return 1
}
