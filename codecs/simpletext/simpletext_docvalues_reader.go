package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.DocValuesProducer = &SimpleTextDocValuesReader{}

type SimpleTextDocValuesReader struct {
}

func NewSimpleTextDocValuesReader(state *index.SegmentReadState, ext string) (*SimpleTextDocValuesReader, error) {
	panic("")
}

func (s *SimpleTextDocValuesReader) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextDocValuesReader) GetNumeric(field *types.FieldInfo) (index.NumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextDocValuesReader) GetBinary(field *types.FieldInfo) (index.BinaryDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextDocValuesReader) GetSorted(field *types.FieldInfo) (index.SortedDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextDocValuesReader) GetSortedNumeric(field *types.FieldInfo) (index.SortedNumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextDocValuesReader) GetSortedSet(field *types.FieldInfo) (index.SortedSetDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextDocValuesReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}