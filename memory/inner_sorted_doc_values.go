package memory

import (
	"github.com/geange/lucene-go/core/index"
)

var _ index.SortedDocValues = &innerSortedDocValues{}

type innerSortedDocValues struct {
	*index.SortedDocValuesDefault

	value []byte
	it    *memoryDocValuesIterator
}

func (i *innerSortedDocValues) TermsEnum() (index.TermsEnum, error) {
	return index.NewSortedDocValuesTermsEnum(i), nil
}

func newInnerSortedDocValues(value []byte) *innerSortedDocValues {
	values := &innerSortedDocValues{
		SortedDocValuesDefault: nil,
		value:                  value,
		it:                     newMemoryDocValuesIterator(),
	}
	values.SortedDocValuesDefault = index.NewSortedDocValuesDefault(&index.SortedDocValuesDefaultConfig{
		OrdValue:      values.OrdValue,
		LookupOrd:     values.LookupOrd,
		GetValueCount: values.GetValueCount,
	})
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
