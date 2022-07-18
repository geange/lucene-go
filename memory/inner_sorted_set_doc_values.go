package memory

import "github.com/geange/lucene-go/core/util"

type innerSortedSetDocValues struct {
	ord      int64
	values   *util.BytesRefHash
	bytesIds []int
	it       *memoryDocValuesIterator
}

func newInnerSortedSetDocValues(values *util.BytesRefHash, bytesIds []int, it *memoryDocValuesIterator) *innerSortedSetDocValues {
	return &innerSortedSetDocValues{values: values, bytesIds: bytesIds, it: it}
}

func (i *innerSortedSetDocValues) DocID() int {
	return i.it.docId()
}

func (i *innerSortedSetDocValues) NextDoc() (int, error) {
	return i.it.nextDoc(), nil
}

func (i *innerSortedSetDocValues) Advance(target int) (int, error) {
	return i.it.advance(target), nil
}

func (i *innerSortedSetDocValues) SlowAdvance(target int) (int, error) {
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

func (i *innerSortedSetDocValues) Cost() int64 {
	return 1
}

func (i *innerSortedSetDocValues) AdvanceExact(target int) (bool, error) {
	i.ord = 0

	advance, err := i.Advance(target)
	if err != nil {
		return false, err
	}
	return advance == target, nil
}

func (i *innerSortedSetDocValues) NextOrd() (int64, error) {
	if int(i.ord) >= i.values.Size() {
		return -1, nil
	}
	ord := i.ord
	i.ord++
	return ord, nil
}

func (i *innerSortedSetDocValues) LookupOrd(ord int64) ([]byte, error) {
	return i.values.Get(i.bytesIds[int(ord)]), nil
}

func (i *innerSortedSetDocValues) GetValueCount() int64 {
	return int64(i.values.Size())
}
