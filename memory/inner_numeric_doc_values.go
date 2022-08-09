package memory

import (
	"github.com/geange/lucene-go/core/index"
)

var _ index.NumericDocValues = &innerNumericDocValues{}

type innerNumericDocValues struct {
	it    *memoryDocValuesIterator
	value int64
}

func newInnerNumericDocValues(value int64) *innerNumericDocValues {
	return &innerNumericDocValues{
		it:    newMemoryDocValuesIterator(),
		value: value,
	}
}

func (i *innerNumericDocValues) DocID() int {
	return i.it.docId()
}

func (i *innerNumericDocValues) NextDoc() (int, error) {
	return i.it.nextDoc(), nil
}

func (i *innerNumericDocValues) Advance(target int) (int, error) {
	return i.it.advance(target), nil
}

func (i *innerNumericDocValues) SlowAdvance(target int) (int, error) {
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

func (i *innerNumericDocValues) Cost() int64 {
	return 1
}

func (i *innerNumericDocValues) AdvanceExact(target int) (bool, error) {
	advance, err := i.Advance(target)
	if err != nil {
		return false, err
	}
	return advance == target, nil
}

func (i *innerNumericDocValues) LongValue() (int64, error) {
	return i.value, nil
}
