package lucene80

import (
	"context"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.DocValuesProducer = &DocValuesProducer{}

type DocValuesProducer struct{}

func NewDocValuesProducer(ctx context.Context, state *index.SegmentReadState,
	dataCodec, dataExtension, metaCodec, metaExtension string) (*DocValuesProducer, error) {
	panic("implement me")
}

func (d *DocValuesProducer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) GetNumeric(ctx context.Context, field *document.FieldInfo) (index.NumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) GetBinary(ctx context.Context, field *document.FieldInfo) (index.BinaryDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) GetSorted(ctx context.Context, fieldInfo *document.FieldInfo) (index.SortedDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) GetSortedNumeric(ctx context.Context, field *document.FieldInfo) (index.SortedNumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) GetSortedSet(ctx context.Context, field *document.FieldInfo) (index.SortedSetDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesProducer) GetMergeInstance() index.DocValuesProducer {
	//TODO implement me
	panic("implement me")
}
