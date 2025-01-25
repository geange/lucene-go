package memory

import (
	"context"
	"io"

	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util/bytesref"
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
	doc, _ := i.iterator.docId()
	return doc
}

func (i *numericDocValues) NextDoc(context.Context) (int, error) {
	return i.iterator.nextDoc()
}

func (i *numericDocValues) Advance(ctx context.Context, target int) (int, error) {
	return i.iterator.advance(target)
}

func (i *numericDocValues) SlowAdvance(ctx context.Context, target int) (int, error) {
	doc := 0
	var err error
	for doc < target {
		doc, err = i.NextDoc(ctx)
		if err != nil {
			return 0, err
		}
	}
	return doc, nil
}

func (i *numericDocValues) Cost() int64 {
	return 1
}

func (i *numericDocValues) AdvanceExact(target int) (bool, error) {
	advance, err := i.Advance(nil, target)
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
	*coreIndex.BaseSortedDocValues

	value []byte
	it    *docValuesIterator
}

func (i *sortedDocValues) TermsEnum() (index.TermsEnum, error) {
	return coreIndex.NewSortedDocValuesTermsEnum(i), nil
}

func newSortedDocValues(value []byte) *sortedDocValues {
	values := &sortedDocValues{
		BaseSortedDocValues: nil,
		value:               value,
		it:                  newDocValuesIterator(),
	}
	values.BaseSortedDocValues = coreIndex.NewBaseSortedDocValues(&coreIndex.SortedDocValuesDefaultConfig{
		OrdValue:      values.OrdValue,
		LookupOrd:     values.LookupOrd,
		GetValueCount: values.GetValueCount,
	})
	return values
}

func (i *sortedDocValues) DocID() int {
	doc, _ := i.it.docId()
	return doc
}

func (i *sortedDocValues) NextDoc(context.Context) (int, error) {
	return i.it.nextDoc()
}

func (i *sortedDocValues) Advance(ctx context.Context, target int) (int, error) {
	return i.it.advance(target)
}

func (i *sortedDocValues) SlowAdvance(ctx context.Context, target int) (int, error) {
	doc := 0
	var err error
	for doc < target {
		doc, err = i.NextDoc(nil)
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
	advance, err := i.Advance(nil, target)
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
	doc, _ := i.it.docId()
	return doc
}

func (i *sortedNumericDocValues) NextDoc(context.Context) (int, error) {
	return i.it.nextDoc()
}

func (i *sortedNumericDocValues) Advance(ctx context.Context, target int) (int, error) {
	return i.it.advance(target)
}

func (i *sortedNumericDocValues) SlowAdvance(ctx context.Context, target int) (int, error) {
	doc := 0
	var err error
	for doc < target {
		doc, err = i.NextDoc(nil)
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
	advance, err := i.Advance(nil, target)
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
	values   *bytesref.BytesHash
	bytesIds []int
	it       *docValuesIterator
}

func newSortedSetDocValues(values *bytesref.BytesHash, bytesIds []int, it *docValuesIterator) *sortedSetDocValues {
	return &sortedSetDocValues{values: values, bytesIds: bytesIds, it: it}
}

func (i *sortedSetDocValues) DocID() int {
	doc, _ := i.it.docId()
	return doc
}

func (i *sortedSetDocValues) NextDoc(context.Context) (int, error) {
	return i.it.nextDoc()
}

func (i *sortedSetDocValues) Advance(ctx context.Context, target int) (int, error) {
	return i.it.advance(target)
}

func (i *sortedSetDocValues) SlowAdvance(ctx context.Context, target int) (int, error) {
	doc := 0
	var err error
	for doc < target {
		doc, err = i.NextDoc(nil)
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

	advance, err := i.Advance(context.Background(), target)
	if err != nil {
		return false, err
	}
	return advance == target, nil
}

func (i *sortedSetDocValues) NextOrd() (int64, error) {
	if int(i.ord) >= i.values.Size() {
		return 0, io.EOF
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
