package index

import (
	"context"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ DocValuesProducer = &EmptyDocValuesProducer{}

type EmptyDocValuesProducer struct {
	FnGetNumeric       func(ctx context.Context, field *document.FieldInfo) (index.NumericDocValues, error)
	FnGetBinary        func(ctx context.Context, field *document.FieldInfo) (index.BinaryDocValues, error)
	FnGetSorted        func(ctx context.Context, field *document.FieldInfo) (index.SortedDocValues, error)
	FnGetSortedNumeric func(ctx context.Context, field *document.FieldInfo) (index.SortedNumericDocValues, error)
	FnGetSortedSet     func(ctx context.Context, field *document.FieldInfo) (index.SortedSetDocValues, error)
	FnCheckIntegrity   func() error
}

func (e *EmptyDocValuesProducer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (e *EmptyDocValuesProducer) GetNumeric(ctx context.Context, field *document.FieldInfo) (index.NumericDocValues, error) {
	return e.FnGetNumeric(ctx, field)
}

func (e *EmptyDocValuesProducer) GetBinary(ctx context.Context, field *document.FieldInfo) (index.BinaryDocValues, error) {
	return e.FnGetBinary(ctx, field)
}

func (e *EmptyDocValuesProducer) GetSorted(ctx context.Context, fieldInfo *document.FieldInfo) (index.SortedDocValues, error) {
	return e.FnGetSorted(ctx, fieldInfo)
}

func (e *EmptyDocValuesProducer) GetSortedNumeric(ctx context.Context, field *document.FieldInfo) (index.SortedNumericDocValues, error) {
	return e.FnGetSortedNumeric(ctx, field)
}

func (e *EmptyDocValuesProducer) GetSortedSet(ctx context.Context, field *document.FieldInfo) (index.SortedSetDocValues, error) {
	return e.FnGetSortedSet(ctx, field)
}

func (e *EmptyDocValuesProducer) CheckIntegrity() error {
	return e.FnCheckIntegrity()
}

func (e *EmptyDocValuesProducer) GetMergeInstance() DocValuesProducer {
	return e
}
