package index

import (
	"errors"
	"fmt"
	"math"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/packed"
)

const (
	HAS_VALUE_MASK    = 1
	HAS_NO_VALUE_MASK = 0
	SHIFT             = 1
)

// DocValuesFieldUpdates
// holds updates of a single docvalues field, for a set of documents within one segment.
type DocValuesFieldUpdates interface {
	Field() string
	AddInt64(doc int, value int64) error
	AddBytes(doc int, value []byte) error

	// AddIterator
	// Adds the item for the given docID. This method prevents conditional calls to
	// DocValuesFieldUpdates.Iterator.longValue() or DocValuesFieldUpdates.Iterator.binaryValue()
	// since the implementation knows if it's a long item iterator or binary item
	AddIterator(doc int, it DocValuesFieldUpdatesIterator) error

	// Iterator
	// Returns an DocValuesFieldUpdates.Iterator over the updated documents and their values.
	Iterator() (DocValuesFieldUpdatesIterator, error)

	// Finish
	// Freezes internal data structures and sorts updates by docID for efficient iteration.
	Finish() error

	// Any
	// Returns true if this instance contains any updates.
	Any() bool

	Size() int

	// Reset
	// Adds an update that resets the documents item.
	// Params: doc – the doc to update
	Reset(doc int) error

	Swap(i, j int) error

	Grow(i int) error

	Resize(i int) error

	EnsureFinished() error
	GetFinished() bool
}

type BaseDocValuesFieldUpdates struct {
	field        string
	_type        document.DocValuesType
	delGen       int64
	bitsPerValue int
	finished     bool
	maxDoc       int
	docs         *packed.FixSizePagedMutable
	size         int
}

func (d *BaseDocValuesFieldUpdates) Field() string {
	return d.field
}

func (d *BaseDocValuesFieldUpdates) Finish() error {
	if d.finished {
		return errors.New("already finished")
	}
	d.finished = true
	// shrink wrap
	if d.size < d.docs.Size() {
		if err := d.Resize(d.size); err != nil {
			return err
		}
	}
	if d.size > 0 {
		// We need a stable sort but InPlaceMergeSorter performs lots of swaps
		// which hurts performance due to all the packed ints we are using.
		// Another option would be TimSorter, but it needs additional API (copy to
		// temp storage, compare with item in temp storage, etc.) so we instead
		// use quicksort and record ords of each update to guarantee stability.
		bitsRequired, err := packed.BitsRequired(int64(d.size - 1))
		if err != nil {
			return err
		}

		ords := packed.DefaultGetMutable(d.size, bitsRequired, packed.DEFAULT)
		for i := 0; i < d.size; i++ {
			ords.Set(i, uint64(i))
		}

	}

	return nil
}

// Any Returns true if this instance contains any updates.
func (d *BaseDocValuesFieldUpdates) Any() bool {
	return d.size > 0
}

func (d *BaseDocValuesFieldUpdates) Size() int {
	return d.size
}

func (d *BaseDocValuesFieldUpdates) Swap(i, j int) error {
	doc1, err := d.docs.Get(j)
	if err != nil {
		return err
	}
	doc2, err := d.docs.Get(i)
	if err != nil {
		return err
	}
	d.docs.Set(j, doc2)
	d.docs.Set(i, doc1)
	return nil
}

func (d *BaseDocValuesFieldUpdates) Grow(size int) error {
	d.docs = d.docs.Grow(size).(*packed.FixSizePagedMutable)
	return nil
}

func (d *BaseDocValuesFieldUpdates) Resize(size int) error {
	docs, ok := d.docs.Resize(size).(*packed.FixSizePagedMutable)
	if !ok {
		return fmt.Errorf("type is not *packed.FixSizePagedMutable")
	}
	d.docs = docs
	return nil
}

func (d *BaseDocValuesFieldUpdates) GetFinished() bool {
	return d.finished
}

func (b *BinaryDocValuesFieldUpdates) add(doc int) (int, error) {
	return b.addInternal(doc, HAS_VALUE_MASK)
}

func (b *BinaryDocValuesFieldUpdates) addInternal(doc int, hasValueMask int64) (int, error) {
	if b.finished {
		return 0, errors.New("already finished")
	}

	if doc >= b.maxDoc {
		return 0, errors.New("doc too big")
	}

	// TODO: if the Sorter interface changes to take long indexes, we can remove that limitation
	if b.size == math.MaxInt32 {
		return 0, errors.New("cannot support more than Integer.MAX_VALUE doc/item entries")
	}

	// grow the structures to have room for more elements
	if b.docs.Size() == b.size {
		if err := b.Grow(b.size + 1); err != nil {
			return 0, err
		}
	}

	value := (int64(doc) << SHIFT) | hasValueMask
	b.docs.Set(b.size, uint64(value))
	b.size++
	return b.size - 1, nil
}

// DocValuesFieldUpdatesIterator
// An iterator over documents and their updated values. Only documents with updates are returned
// by this iterator, and the documents are returned in increasing order.
type DocValuesFieldUpdatesIterator interface {
	types.DocValuesIterator

	// LongValue
	// Returns a long item for the current document if this iterator is a long iterator.
	LongValue() (int64, error)

	// BinaryValue
	// Returns a binary item for the current document if this iterator is a binary item iterator.
	BinaryValue() ([]byte, error)

	// DelGen
	// Returns delGen for this packet.
	DelGen() int64

	// HasValue
	// Returns true if this doc has a item
	HasValue() bool
}

type DVFUIterator struct {
}

func (*DVFUIterator) AdvanceExact(target int) (bool, error) {
	return false, errors.New("unsupported operation exception")
}

func (*DVFUIterator) Advance(target int) (int, error) {
	return 0, errors.New("unsupported operation exception")
}

func (*DVFUIterator) Cost() int64 {
	return 0
}

func AsBinaryDocValues(iterator DocValuesFieldUpdatesIterator) index.BinaryDocValues {
	return &BaseBinaryDocValues{
		FnDocID:        iterator.DocID,
		FnNextDoc:      iterator.NextDoc,
		FnAdvance:      iterator.Advance,
		FnSlowAdvance:  iterator.SlowAdvance,
		FnCost:         iterator.Cost,
		FnAdvanceExact: iterator.AdvanceExact,
		FnBinaryValue:  iterator.BinaryValue,
	}
}

func AsNumericDocValues(iterator DocValuesFieldUpdatesIterator) index.NumericDocValues {
	return &NumericDocValuesDefault{
		FnDocID:        iterator.DocID,
		FnNextDoc:      iterator.NextDoc,
		FnAdvance:      iterator.Advance,
		FnSlowAdvance:  iterator.SlowAdvance,
		FnCost:         iterator.Cost,
		FnAdvanceExact: iterator.AdvanceExact,
		FnLongValue:    iterator.LongValue,
	}
}

type SingleValueDocValuesFieldUpdates struct {
}
