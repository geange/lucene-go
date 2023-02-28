package index

import (
	"errors"
	"fmt"
	"github.com/geange/lucene-go/core/types"
	"io"
)

var _ DocValuesWriter = &BinaryDocValuesWriter{}

type BinaryDocValuesWriter struct {
	bytes         [][]byte
	docsWithField *DocsWithFieldSet
	fieldInfo     *types.FieldInfo
	lastDocID     int
}

func NewBinaryDocValuesWriter(fieldInfo *types.FieldInfo) *BinaryDocValuesWriter {
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
	return consumer.AddBinaryField(b.fieldInfo, &EmptyDocValuesProducer{
		FnGetBinary: func(field *types.FieldInfo) (BinaryDocValues, error) {
			iterator, err := b.docsWithField.Iterator()
			if err != nil {
				return nil, err
			}
			return NewBufferedBinaryDocValues(b.bytes, iterator), nil
		},
	})
}

func (b *BinaryDocValuesWriter) GetDocValues() DocIdSetIterator {
	iterator, _ := b.docsWithField.Iterator()
	return NewBufferedBinaryDocValues(b.bytes, iterator)
}

var _ BinaryDocValues = &BufferedBinaryDocValues{}

type BufferedBinaryDocValues struct {
	*DocIdSetIteratorDefault
	docsWithField DocIdSetIterator
	values        [][]byte
	pos           int
}

func NewBufferedBinaryDocValues(values [][]byte, docsWithField DocIdSetIterator) *BufferedBinaryDocValues {
	docValues := &BufferedBinaryDocValues{
		docsWithField: docsWithField,
		values:        values,
		pos:           -1,
	}
	docValues.DocIdSetIteratorDefault = NewDocIdSetIteratorDefault(
		&DocIdSetIteratorDefaultConfig{
			NextDoc: docValues.NextDoc,
		},
	)
	return docValues
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
