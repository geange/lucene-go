package index

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/bytesref"
	"github.com/geange/lucene-go/core/util/packed"
)

type BaseBinaryDocValues struct {
	FnDocID        func() int
	FnNextDoc      func(ctx context.Context) (int, error)
	FnAdvance      func(ctx context.Context, target int) (int, error)
	FnSlowAdvance  func(ctx context.Context, target int) (int, error)
	FnCost         func() int64
	FnAdvanceExact func(target int) (bool, error)
	FnBinaryValue  func() ([]byte, error)
}

func (n *BaseBinaryDocValues) DocID() int {
	return n.FnDocID()
}

func (n *BaseBinaryDocValues) NextDoc(ctx context.Context) (int, error) {
	return n.FnNextDoc(ctx)
}

func (n *BaseBinaryDocValues) Advance(ctx context.Context, target int) (int, error) {
	return n.FnAdvance(ctx, target)
}

func (n *BaseBinaryDocValues) SlowAdvance(ctx context.Context, target int) (int, error) {
	if n.FnSlowAdvance != nil {
		return n.FnSlowAdvance(ctx, target)
	}
	return types.SlowAdvanceWithContext(ctx, n, target)
}

func (n *BaseBinaryDocValues) Cost() int64 {
	return n.FnCost()
}

func (n *BaseBinaryDocValues) AdvanceExact(target int) (bool, error) {
	return n.FnAdvanceExact(target)
}

func (n *BaseBinaryDocValues) BinaryValue() ([]byte, error) {
	return n.FnBinaryValue()
}

var _ DocValuesFieldUpdates = &BinaryDocValuesFieldUpdates{}

// BinaryDocValuesFieldUpdates
// A DocValuesFieldUpdates which holds updates of documents, of a single BinaryDocValuesField.
// lucene.experimental
type BinaryDocValuesFieldUpdates struct {
	*BaseDocValuesFieldUpdates

	offsets, lengths *packed.PagedGrowableWriter

	values bytesref.Builder
}

func (b *BinaryDocValuesFieldUpdates) AddInt64(doc int, value int64) error {
	return errors.New("unsupported operation exception")
}

func (b *BinaryDocValuesFieldUpdates) AddBytes(doc int, value []byte) error {
	idx, err := b.add(doc)
	if err != nil {
		return err
	}
	b.offsets.Set(idx, uint64(b.values.Length()))
	b.lengths.Set(idx, uint64(len(value)))
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
	if err := b.BaseDocValuesFieldUpdates.Swap(i, j); err != nil {
		return err
	}

	tmpOffset, err := b.offsets.Get(j)
	if err != nil {
		return err
	}
	v1, err := b.offsets.Get(i)
	if err != nil {
		return err
	}
	b.offsets.Set(j, v1)
	b.offsets.Set(i, tmpOffset)

	tmpLength, err := b.lengths.Get(j)
	if err != nil {
		return err
	}
	v2, err := b.lengths.Get(i)
	if err != nil {
		return err
	}
	b.lengths.Set(j, v2)
	b.lengths.Set(i, tmpLength)
	return nil
}

func (b *BinaryDocValuesFieldUpdates) Grow(size int) error {
	if err := b.BaseDocValuesFieldUpdates.Grow(size); err != nil {
		return err
	}
	b.offsets = b.offsets.Grow(size).(*packed.PagedGrowableWriter)
	b.lengths = b.lengths.Grow(size).(*packed.PagedGrowableWriter)
	return nil
}

func (b *BinaryDocValuesFieldUpdates) Resize(size int) error {
	if err := b.BaseDocValuesFieldUpdates.Resize(size); err != nil {
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

func (b *BinaryDocValuesWriter) Flush(state *index.SegmentWriteState, sortMap index.DocMap, consumer index.DocValuesConsumer) error {
	return consumer.AddBinaryField(context.TODO(), b.fieldInfo, &EmptyDocValuesProducer{
		FnGetBinary: func(ctx context.Context, field *document.FieldInfo) (index.BinaryDocValues, error) {
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

var _ index.BinaryDocValues = &BufferedBinaryDocValues{}

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

func (b *BufferedBinaryDocValues) NextDoc(ctx context.Context) (int, error) {
	doc, err := b.docsWithField.NextDoc(ctx)
	if err != nil {
		return 0, err
	}
	b.pos++
	return doc, nil
}

func (b *BufferedBinaryDocValues) Advance(ctx context.Context, target int) (int, error) {
	return 0, errors.New("unsupported operation exception")
}

func (b *BufferedBinaryDocValues) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvanceWithContext(ctx, b, target)
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
