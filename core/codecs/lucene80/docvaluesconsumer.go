package lucene80

import (
	"context"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.DocValuesConsumer = &DocValuesConsumer{}

type DocValuesConsumer struct{}

type DocValuesConsumerMode byte

const (
	BEST_SPEED = DocValuesConsumerMode(iota)
	BEST_COMPRESSION
)

func NewDocValuesConsumer(ctx context.Context, state *index.SegmentWriteState,
	dataCodec, dataExtension, metaCodec, metaExtension string, mode DocValuesConsumerMode) (*DocValuesConsumer, error) {
	panic("implement me")
}

func (d *DocValuesConsumer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesConsumer) AddNumericField(ctx context.Context, field *document.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesConsumer) AddBinaryField(ctx context.Context, field *document.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesConsumer) AddSortedField(ctx context.Context, field *document.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesConsumer) AddSortedNumericField(ctx context.Context, field *document.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesConsumer) AddSortedSetField(ctx context.Context, field *document.FieldInfo, valuesProducer index.DocValuesProducer) error {
	//TODO implement me
	panic("implement me")
}
