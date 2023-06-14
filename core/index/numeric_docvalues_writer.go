package index

import (
	"errors"
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/packed"
	"io"
)

var _ DocValuesWriter = &NumericDocValuesWriter{}

type NumericDocValuesWriter struct {
	pending       *packed.PackedLongValuesBuilder
	finalValues   *packed.PackedLongValues
	docsWithField *DocsWithFieldSet
	fieldInfo     *types.FieldInfo
	lastDocID     int
}

func NewNumericDocValuesWriter(fieldInfo *types.FieldInfo) *NumericDocValuesWriter {
	panic("")
}

func (n *NumericDocValuesWriter) AddValue(docID int, value int64) error {
	if docID <= n.lastDocID {
		panic("")
	}
	n.pending.Add(value)
	if err := n.docsWithField.Add(docID); err != nil {
		return err
	}
	n.lastDocID = docID
	return nil
}

func (n *NumericDocValuesWriter) Flush(state *SegmentWriteState, sortMap DocMap, consumer DocValuesConsumer) error {
	n.finalValues = n.pending.Build()

	return consumer.AddNumericField(n.fieldInfo, &EmptyDocValuesProducer{
		FnGetNumeric: func(field *types.FieldInfo) (NumericDocValues, error) {
			iterator, err := n.docsWithField.Iterator()
			if err != nil {
				return nil, err
			}
			return NewBufferedNumericDocValues(n.finalValues, iterator), nil
		},
	})
}

func (n *NumericDocValuesWriter) GetDocValues() DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}

var _ NumericDocValues = &BufferedNumericDocValues{}

type BufferedNumericDocValues struct {
	iter          *packed.PackedLongValuesIterator
	docsWithField DocIdSetIterator
	value         int64
}

func NewBufferedNumericDocValues(values *packed.PackedLongValues,
	docsWithFields DocIdSetIterator) *BufferedNumericDocValues {

	docValues := &BufferedNumericDocValues{
		iter:          values.Iterator(),
		docsWithField: docsWithFields,
		value:         0,
	}
	return docValues
}

func (b *BufferedNumericDocValues) DocID() int {
	return b.docsWithField.DocID()
}

func (b *BufferedNumericDocValues) NextDoc() (int, error) {
	docID, err := b.docsWithField.NextDoc()
	if err != nil {
		return 0, err
	}
	b.value = b.iter.Next()
	return docID, nil
}

func (b *BufferedNumericDocValues) Advance(target int) (int, error) {
	return 0, errors.New("unsupported Operation")
}

func (b *BufferedNumericDocValues) SlowAdvance(target int) (int, error) {
	return SlowAdvance(b, target)
}

func (b *BufferedNumericDocValues) Cost() int64 {
	return b.docsWithField.Cost()
}

func (b *BufferedNumericDocValues) AdvanceExact(target int) (bool, error) {
	return false, errors.New("unsupported Operation")
}

func (b *BufferedNumericDocValues) LongValue() (int64, error) {
	return b.value, nil
}

var _ NumericDocValues = &SortingNumericDocValues{}

type SortingNumericDocValues struct {
	dvs   *NumericDVs
	docID int
	cost  int
}

func (s *SortingNumericDocValues) DocID() int {
	return s.docID
}

func (s *SortingNumericDocValues) NextDoc() (int, error) {
	value, ok := s.dvs.docsWithField.NextSet(uint(s.docID + 1))
	if !ok {
		return 0, io.EOF
	}
	s.docID = int(value)
	return s.docID, nil
}

func (s *SortingNumericDocValues) Advance(target int) (int, error) {
	return 0, errors.New("unsupported Operation")
}

func (s *SortingNumericDocValues) SlowAdvance(target int) (int, error) {
	return SlowAdvance(s, target)
}

func (s *SortingNumericDocValues) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (s *SortingNumericDocValues) AdvanceExact(target int) (bool, error) {
	s.docID = target
	return s.dvs.docsWithField.Test(uint(target)), nil
}

func (s *SortingNumericDocValues) LongValue() (int64, error) {
	return s.dvs.values[s.docID], nil
}

type NumericDVs struct {
	values        []int64
	docsWithField *bitset.BitSet
}

func NewNumericDVs(values []int64, docsWithField *bitset.BitSet) *NumericDVs {
	return &NumericDVs{values: values, docsWithField: docsWithField}
}
