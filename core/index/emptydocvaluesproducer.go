package index

import (
	"context"
	"github.com/geange/lucene-go/core/document"
)

var _ DocValuesProducer = &EmptyDocValuesProducer{}

type EmptyDocValuesProducer struct {
	FnGetNumeric       func(ctx context.Context, field *document.FieldInfo) (NumericDocValues, error)
	FnGetBinary        func(ctx context.Context, field *document.FieldInfo) (BinaryDocValues, error)
	FnGetSorted        func(ctx context.Context, field *document.FieldInfo) (SortedDocValues, error)
	FnGetSortedNumeric func(ctx context.Context, field *document.FieldInfo) (SortedNumericDocValues, error)
	FnGetSortedSet     func(ctx context.Context, field *document.FieldInfo) (SortedSetDocValues, error)
	FnCheckIntegrity   func() error
}

func (e *EmptyDocValuesProducer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) GetNumeric(ctx context.Context, field *document.FieldInfo) (NumericDocValues, error) {
	return e.FnGetNumeric(ctx, field)
}

func (e *EmptyDocValuesProducer) GetBinary(ctx context.Context, field *document.FieldInfo) (BinaryDocValues, error) {
	return e.FnGetBinary(ctx, field)
}

func (e *EmptyDocValuesProducer) GetSorted(ctx context.Context, fieldInfo *document.FieldInfo) (SortedDocValues, error) {
	return e.FnGetSorted(ctx, fieldInfo)
}

func (e *EmptyDocValuesProducer) GetSortedNumeric(ctx context.Context, field *document.FieldInfo) (SortedNumericDocValues, error) {
	return e.FnGetSortedNumeric(ctx, field)
}

func (e *EmptyDocValuesProducer) GetSortedSet(ctx context.Context, field *document.FieldInfo) (SortedSetDocValues, error) {
	return e.FnGetSortedSet(ctx, field)
}

func (e *EmptyDocValuesProducer) CheckIntegrity() error {
	return e.FnCheckIntegrity()
}

func (e *EmptyDocValuesProducer) GetMergeInstance() DocValuesProducer {
	return e
}
