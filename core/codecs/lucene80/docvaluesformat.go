package lucene80

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.DocValuesFormat = &DocValuesFormat{}

type DocValuesFormat struct {
}

func NewDocValuesFormat() *DocValuesFormat {
	return &DocValuesFormat{}
}

func (d *DocValuesFormat) GetName() string {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesFormat) FieldsConsumer(ctx context.Context, state *index.SegmentWriteState) (index.DocValuesConsumer, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesFormat) FieldsProducer(ctx context.Context, state *index.SegmentReadState) (index.DocValuesProducer, error) {
	//TODO implement me
	panic("implement me")
}
