package simpletext

import (
	"bytes"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

var (
	DOC_VALUES_END   = []byte("END")
	DOC_VALUES_FIELD = []byte("field ")
	DOC_VALUES_TYPE  = []byte("  type ")
	// used for numerics
	DOC_VALUES_MINVALUE = []byte("  minvalue ")
	DOC_VALUES_PATTERN  = []byte("  pattern ")
	// used for bytes
	DOC_VALUES_LENGTH    = []byte("length ")
	DOC_VALUES_MAXLENGTH = []byte("  maxlength ")
	// used for sorted bytes
	DOC_VALUES_NUMVALUES  = []byte("  numvalues ")
	DOC_VALUES_ORDPATTERN = []byte("  ordpattern ")
)

var _ index.DocValuesConsumer = &SimpleTextDocValuesWriter{}

type SimpleTextDocValuesWriter struct {
	data       store.IndexOutput
	scratch    *bytes.Buffer
	fieldsSeen map[string]struct{}
}

func NewSimpleTextDocValuesWriter(state *index.SegmentWriteState, ext string) (*SimpleTextDocValuesWriter, error) {
	panic("")
}

func (s *SimpleTextDocValuesWriter) AddNumericField(field *types.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextDocValuesWriter) AddBinaryField(field *types.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextDocValuesWriter) AddSortedField(field *types.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextDocValuesWriter) AddSortedNumericField(field *types.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextDocValuesWriter) AddSortedSetField(field *types.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}
