package simpletext

import (
	"context"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	index2 "github.com/geange/lucene-go/core/interface/index"
)

var _ index2.NormsFormat = &NormsFormat{}

const (
	NORMS_SEG_EXTENSION = "len"
)

// NormsFormat plain-text norms format.
// FOR RECREATIONAL USE ONLY
type NormsFormat struct {
}

func NewNormsFormat() *NormsFormat {
	return &NormsFormat{}
}

func (s *NormsFormat) NormsConsumer(ctx context.Context, state *index2.SegmentWriteState) (index2.NormsConsumer, error) {
	return NewSimpleTextNormsConsumer(nil, state)
}

func (s *NormsFormat) NormsProducer(ctx context.Context, state *index2.SegmentReadState) (index2.NormsProducer, error) {
	return NewSimpleTextNormsProducer(state)
}

var _ index2.NormsProducer = &SimpleTextNormsProducer{}

type SimpleTextNormsProducer struct {
	impl *DocValuesReader
}

func NewSimpleTextNormsProducer(state *index2.SegmentReadState) (*SimpleTextNormsProducer, error) {
	reader, err := NewDocValuesReader(context.TODO(), state, NORMS_SEG_EXTENSION)
	if err != nil {
		return nil, err
	}
	return &SimpleTextNormsProducer{impl: reader}, nil
}

func (s *SimpleTextNormsProducer) GetNorms(field *document.FieldInfo) (index2.NumericDocValues, error) {
	return s.impl.GetNumeric(nil, field)
}

func (s *SimpleTextNormsProducer) Close() error {
	return s.impl.Close()
}

func (s *SimpleTextNormsProducer) CheckIntegrity() error {
	return s.impl.CheckIntegrity()
}

func (s *SimpleTextNormsProducer) GetMergeInstance() index2.NormsProducer {
	return s
}

var _ index2.NormsConsumer = &SimpleTextNormsConsumer{}

// SimpleTextNormsConsumer
// Writes plain-text norms.
// FOR RECREATIONAL USE ONLY
type SimpleTextNormsConsumer struct {
	*index.NormsConsumerDefault

	impl *DocValuesWriter
}

func NewSimpleTextNormsConsumer(ctx context.Context, state *index2.SegmentWriteState) (*SimpleTextNormsConsumer, error) {
	writer, err := NewDocValuesWriter(ctx, state, NORMS_SEG_EXTENSION)
	if err != nil {
		return nil, err
	}

	consumer := &SimpleTextNormsConsumer{
		impl: writer,
	}
	consumer.NormsConsumerDefault = &index.NormsConsumerDefault{
		FnAddNormsField: consumer.AddNormsField,
	}
	return consumer, nil
}

func (s *SimpleTextNormsConsumer) Close() error {
	return s.impl.Close()
}

func (s *SimpleTextNormsConsumer) AddNormsField(ctx context.Context, field *document.FieldInfo, normsProducer index2.NormsProducer) error {
	producer := struct {
		*index.EmptyDocValuesProducer
	}{}
	producer.FnGetNumeric = func(ctx context.Context, field *document.FieldInfo) (index2.NumericDocValues, error) {
		return normsProducer.GetNorms(field)
	}

	return s.impl.AddNumericField(ctx, field, producer)
}
