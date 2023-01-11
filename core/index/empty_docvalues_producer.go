package index

import "github.com/geange/lucene-go/core/types"

var _ DocValuesProducer = &EmptyDocValuesProducer{}

type EmptyDocValuesProducer struct {
	FnGetNumeric       func(field *types.FieldInfo) (NumericDocValues, error)
	FnGetBinary        func(field *types.FieldInfo) (BinaryDocValues, error)
	FnGetSorted        func(field *types.FieldInfo) (SortedDocValues, error)
	FnGetSortedNumeric func(field *types.FieldInfo) (SortedNumericDocValues, error)
	FnGetSortedSet     func(field *types.FieldInfo) (SortedSetDocValues, error)
	FnCheckIntegrity   func() error
}

func (e *EmptyDocValuesProducer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) GetNumeric(field *types.FieldInfo) (NumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) GetBinary(field *types.FieldInfo) (BinaryDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) GetSorted(field *types.FieldInfo) (SortedDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) GetSortedNumeric(field *types.FieldInfo) (SortedNumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) GetSortedSet(field *types.FieldInfo) (SortedSetDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}
