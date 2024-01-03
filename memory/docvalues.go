package memory

import (
	"io"

	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util/bytesutils"
)

var _ index.NumericDocValues = &numericDocValues{}

type numericDocValues struct {
	iterator *docValuesIterator
	value    int64
}

func newNumericDocValues(value int64) *numericDocValues {
	return &numericDocValues{
		iterator: newDocValuesIterator(),
		value:    value,
	}
}

func (i *numericDocValues) DocID() int {
	return i.iterator.docId()
}

func (i *numericDocValues) NextDoc() (int, error) {
	return i.iterator.nextDoc(), nil
}

func (i *numericDocValues) Advance(target int) (int, error) {
	return i.iterator.advance(target), nil
}

func (i *numericDocValues) SlowAdvance(target int) (int, error) {
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

func (i *numericDocValues) Cost() int64 {
	return 1
}

func (i *numericDocValues) AdvanceExact(target int) (bool, error) {
	advance, err := i.Advance(target)
	if err != nil {
		return false, err
	}
	return advance == target, nil
}

func (i *numericDocValues) LongValue() (int64, error) {
	return i.value, nil
}

//---------------------------------------------------------------------------------
//---------------------------------------------------------------------------------

var _ index.SortedDocValues = &sortedDocValues{}

type sortedDocValues struct {
	*index.BaseSortedDocValues

	value []byte
	it    *docValuesIterator
}

func (i *sortedDocValues) TermsEnum() (index.TermsEnum, error) {
	return index.NewSortedDocValuesTermsEnum(i), nil
}

func newSortedDocValues(value []byte) *sortedDocValues {
	values := &sortedDocValues{
		BaseSortedDocValues: nil,
		value:               value,
		it:                  newDocValuesIterator(),
	}
	values.BaseSortedDocValues = index.NewBaseSortedDocValues(&index.SortedDocValuesDefaultConfig{
		OrdValue:      values.OrdValue,
		LookupOrd:     values.LookupOrd,
		GetValueCount: values.GetValueCount,
	})
	return values
}

func (i *sortedDocValues) DocID() int {
	return i.it.docId()
}

func (i *sortedDocValues) NextDoc() (int, error) {
	return i.it.nextDoc(), nil
}

func (i *sortedDocValues) Advance(target int) (int, error) {
	return i.it.advance(target), nil
}

func (i *sortedDocValues) SlowAdvance(target int) (int, error) {
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

func (i *sortedDocValues) Cost() int64 {
	return 1
}

func (i *sortedDocValues) AdvanceExact(target int) (bool, error) {
	advance, err := i.Advance(target)
	if err != nil {
		return false, err
	}
	return advance == target, nil
}

func (i *sortedDocValues) OrdValue() (int, error) {
	return 0, nil
}

func (i *sortedDocValues) LookupOrd(ord int) ([]byte, error) {
	return i.value, nil
}

func (i *sortedDocValues) GetValueCount() int {
	return 1
}

//---------------------------------------------------------------------------------
//---------------------------------------------------------------------------------

var _ index.SortedNumericDocValues = &sortedNumericDocValues{}

type sortedNumericDocValues struct {
	it     *docValuesIterator
	ord    int
	values []int
	count  int
}

func newSortedNumericDocValues(values []int, count int) *sortedNumericDocValues {
	return &sortedNumericDocValues{
		it:     newDocValuesIterator(),
		ord:    0,
		values: values,
		count:  count,
	}
}

func (i *sortedNumericDocValues) DocID() int {
	return i.it.docId()
}

func (i *sortedNumericDocValues) NextDoc() (int, error) {
	return i.it.nextDoc(), nil
}

func (i *sortedNumericDocValues) Advance(target int) (int, error) {
	return i.it.advance(target), nil
}

func (i *sortedNumericDocValues) SlowAdvance(target int) (int, error) {
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

func (i *sortedNumericDocValues) Cost() int64 {
	return 1
}

func (i *sortedNumericDocValues) AdvanceExact(target int) (bool, error) {
	advance, err := i.Advance(target)
	if err != nil {
		return false, err
	}
	return advance == target, nil
}

func (i *sortedNumericDocValues) NextValue() (int64, error) {
	ord := i.ord
	i.ord++
	return int64(i.values[ord]), nil
}

func (i *sortedNumericDocValues) DocValueCount() int {
	return i.count
}

//---------------------------------------------------------------------------------
//---------------------------------------------------------------------------------

var _ index.SortedSetDocValues = &sortedSetDocValues{}

type sortedSetDocValues struct {
	ord      int64
	values   *bytesutils.BytesHash
	bytesIds []int
	it       *docValuesIterator
}

func newSortedSetDocValues(values *bytesutils.BytesHash, bytesIds []int, it *docValuesIterator) *sortedSetDocValues {
	return &sortedSetDocValues{values: values, bytesIds: bytesIds, it: it}
}

func (i *sortedSetDocValues) DocID() int {
	return i.it.docId()
}

func (i *sortedSetDocValues) NextDoc() (int, error) {
	return i.it.nextDoc(), nil
}

func (i *sortedSetDocValues) Advance(target int) (int, error) {
	return i.it.advance(target), nil
}

func (i *sortedSetDocValues) SlowAdvance(target int) (int, error) {
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

func (i *sortedSetDocValues) Cost() int64 {
	return 1
}

func (i *sortedSetDocValues) AdvanceExact(target int) (bool, error) {
	i.ord = 0

	advance, err := i.Advance(target)
	if err != nil {
		return false, err
	}
	return advance == target, nil
}

func (i *sortedSetDocValues) NextOrd() (int64, error) {
	if int(i.ord) >= i.values.Size() {
		return -1, io.EOF
	}
	ord := i.ord
	i.ord++
	return ord, nil
}

func (i *sortedSetDocValues) LookupOrd(ord int64) ([]byte, error) {
	return i.values.Get(i.bytesIds[int(ord)]), nil
}

func (i *sortedSetDocValues) GetValueCount() int64 {
	return int64(i.values.Size())
}
