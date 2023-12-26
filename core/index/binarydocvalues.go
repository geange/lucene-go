package index

import (
	"errors"
	"fmt"
	"io"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/bytesutils"
	"github.com/geange/lucene-go/core/util/packed"
)

// BinaryDocValues A per-document numeric item.
type BinaryDocValues interface {
	types.DocValuesIterator

	// BinaryValue Returns the binary item for the current document ID. It is illegal to call this method after
	// advanceExact(int) returned false.
	// Returns: binary item
	BinaryValue() ([]byte, error)
}

type BinaryDocValuesDefault struct {
	FnDocID        func() int
	FnNextDoc      func() (int, error)
	FnAdvance      func(target int) (int, error)
	FnSlowAdvance  func(target int) (int, error)
	FnCost         func() int64
	FnAdvanceExact func(target int) (bool, error)
	FnBinaryValue  func() ([]byte, error)
}

func (n *BinaryDocValuesDefault) DocID() int {
	return n.FnDocID()
}

func (n *BinaryDocValuesDefault) NextDoc() (int, error) {
	return n.FnNextDoc()
}

func (n *BinaryDocValuesDefault) Advance(target int) (int, error) {
	return n.FnAdvance(target)
}

func (n *BinaryDocValuesDefault) SlowAdvance(target int) (int, error) {
	if n.FnSlowAdvance != nil {
		return n.FnSlowAdvance(target)
	}
	return types.SlowAdvance(n, target)
}

func (n *BinaryDocValuesDefault) Cost() int64 {
	return n.FnCost()
}

func (n *BinaryDocValuesDefault) AdvanceExact(target int) (bool, error) {
	return n.FnAdvanceExact(target)
}

func (n *BinaryDocValuesDefault) BinaryValue() ([]byte, error) {
	return n.FnBinaryValue()
}

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

func (b *BinaryDocValuesFieldUpdates) AddIterator(doc int, it DocValuesFieldUpdatesIterator) error {
	bytes, err := it.BinaryValue()
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
	if err := b.DocValuesFieldUpdatesDefault.Grow(size); err != nil {
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

var _ DocValuesWriter = &BinaryDocValuesWriter{}

type BinaryDocValuesWriter struct {
	bytes         [][]byte
	docsWithField *DocsWithFieldSet
	fieldInfo     *document.FieldInfo
	lastDocID     int
}

func NewBinaryDocValuesWriter(fieldInfo *document.FieldInfo) *BinaryDocValuesWriter {
	return &BinaryDocValuesWriter{
		bytes:         make([][]byte, 0),
		docsWithField: NewDocsWithFieldSet(),
		fieldInfo:     fieldInfo,
		lastDocID:     -1,
	}
}

func (b *BinaryDocValuesWriter) AddValue(docID int, value []byte) error {
	if b.lastDocID >= docID {
		return fmt.Errorf("docID(%d) is small than lastDocID(%d)", docID, b.lastDocID)
	}

	b.lastDocID = docID
	b.bytes = append(b.bytes, value)
	return b.docsWithField.Add(docID)
}

func (b *BinaryDocValuesWriter) Flush(state *SegmentWriteState, sortMap DocMap, consumer DocValuesConsumer) error {
	return consumer.AddBinaryField(nil, b.fieldInfo, &EmptyDocValuesProducer{
		FnGetBinary: func(field *document.FieldInfo) (BinaryDocValues, error) {
			iterator, err := b.docsWithField.Iterator()
			if err != nil {
				return nil, err
			}
			return NewBufferedBinaryDocValues(b.bytes, iterator), nil
		},
	})
}

func (b *BinaryDocValuesWriter) GetDocValues() types.DocIdSetIterator {
	iterator, _ := b.docsWithField.Iterator()
	return NewBufferedBinaryDocValues(b.bytes, iterator)
}

var _ BinaryDocValues = &BufferedBinaryDocValues{}

type BufferedBinaryDocValues struct {
	docsWithField types.DocIdSetIterator
	values        [][]byte
	pos           int
}

func NewBufferedBinaryDocValues(values [][]byte, docsWithField types.DocIdSetIterator) *BufferedBinaryDocValues {
	return &BufferedBinaryDocValues{
		docsWithField: docsWithField,
		values:        values,
		pos:           -1,
	}
}

func (b *BufferedBinaryDocValues) DocID() int {
	return b.docsWithField.DocID()
}

func (b *BufferedBinaryDocValues) NextDoc() (int, error) {
	doc, err := b.docsWithField.NextDoc()
	if err != nil {
		return 0, err
	}
	b.pos++
	return doc, nil
}

func (b *BufferedBinaryDocValues) Advance(target int) (int, error) {
	return 0, errors.New("unsupported operation exception")
}

func (b *BufferedBinaryDocValues) SlowAdvance(target int) (int, error) {
	return types.SlowAdvance(b, target)
}

func (b *BufferedBinaryDocValues) Cost() int64 {
	return 0
}

func (b *BufferedBinaryDocValues) AdvanceExact(target int) (bool, error) {
	return false, errors.New("unsupported operation exception")
}

func (b *BufferedBinaryDocValues) BinaryValue() ([]byte, error) {
	if b.pos >= len(b.values) {
		return nil, io.EOF
	}
	return b.values[b.pos], nil
}
