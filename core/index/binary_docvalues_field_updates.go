package index

import (
	"errors"
	"github.com/geange/lucene-go/core/util/bytesutils"
	"github.com/geange/lucene-go/core/util/packed"
)

var _ DocValuesFieldUpdates = &BinaryDocValuesFieldUpdates{}

// BinaryDocValuesFieldUpdates
// A DocValuesFieldUpdates which holds updates of documents, of a single BinaryDocValuesField.
// lucene.experimental
type BinaryDocValuesFieldUpdates struct {
	*DocValuesFieldUpdatesDefault

	offsets, lengths *packed.PagedGrowableWriter

	values bytesutils.BytesRefBuilder
}

func (b *BinaryDocValuesFieldUpdates) AddInt64(doc int, value int64) error {
	return errors.New("unsupported operation exception")
}

func (b *BinaryDocValuesFieldUpdates) AddBytes(doc int, value []byte) error {
	index, err := b.add(doc)
	if err != nil {
		return err
	}
	b.offsets.Set(index, uint64(b.values.Length()))
	b.lengths.Set(index, uint64(len(value)))
	b.values.AppendBytes(value)
	return nil
}

func (b *BinaryDocValuesFieldUpdates) AddIterator(doc int, value DocValuesFieldUpdatesIterator) error {
	bytes, err := value.BinaryValue()
	if err != nil {
		return err
	}
	return b.AddBytes(doc, bytes)
}

func (b *BinaryDocValuesFieldUpdates) Iterator() (DocValuesFieldUpdatesIterator, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BinaryDocValuesFieldUpdates) Finish() error {
	//TODO implement me
	panic("implement me")
}

func (b *BinaryDocValuesFieldUpdates) Reset(doc int) error {
	//TODO implement me
	panic("implement me")
}

func (b *BinaryDocValuesFieldUpdates) Swap(i, j int) error {
	if err := b.DocValuesFieldUpdatesDefault.Swap(i, j); err != nil {
		return err
	}

	tmpOffset := b.offsets.Get(j)
	b.offsets.Set(j, b.offsets.Get(i))
	b.offsets.Set(i, tmpOffset)

	tmpLength := b.lengths.Get(j)
	b.lengths.Set(j, b.lengths.Get(i))
	b.lengths.Set(i, tmpLength)
	return nil
}

func (b *BinaryDocValuesFieldUpdates) Grow(size int) error {
	err := b.DocValuesFieldUpdatesDefault.Grow(size)
	if err != nil {
		return err
	}
	b.offsets = b.offsets.Grow(size).(*packed.PagedGrowableWriter)
	b.lengths = b.lengths.Grow(size).(*packed.PagedGrowableWriter)
	return nil
}

func (b *BinaryDocValuesFieldUpdates) Resize(size int) error {
	err := b.DocValuesFieldUpdatesDefault.Resize(size)
	if err != nil {
		return err
	}
	b.offsets = b.offsets.Resize(size).(*packed.PagedGrowableWriter)
	b.lengths = b.lengths.Resize(size).(*packed.PagedGrowableWriter)
	return nil
}

func (b *BinaryDocValuesFieldUpdates) EnsureFinished() error {
	//TODO implement me
	panic("implement me")
}
