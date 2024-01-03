package index

import (
	"github.com/geange/lucene-go/core/document"
)

var _ DocValuesProducer = &EmptyDocValuesProducer{}

type EmptyDocValuesProducer struct {
	FnGetNumeric       func(field *document.FieldInfo) (NumericDocValues, error)
	FnGetBinary        func(field *document.FieldInfo) (BinaryDocValues, error)
	FnGetSorted        func(field *document.FieldInfo) (SortedDocValues, error)
	FnGetSortedNumeric func(field *document.FieldInfo) (SortedNumericDocValues, error)
	FnGetSortedSet     func(field *document.FieldInfo) (SortedSetDocValues, error)
	FnCheckIntegrity   func() error
}

func (e *EmptyDocValuesProducer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) GetNumeric(field *document.FieldInfo) (NumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) GetBinary(field *document.FieldInfo) (BinaryDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) GetSorted(field *document.FieldInfo) (SortedDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) GetSortedNumeric(field *document.FieldInfo) (SortedNumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) GetSortedSet(field *document.FieldInfo) (SortedSetDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}
