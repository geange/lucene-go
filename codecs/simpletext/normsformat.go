package simpletext

import (
	"context"

	"github.com/geange/lucene-go/core/document"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.NormsFormat = &NormsFormat{}

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

func (s *NormsFormat) NormsConsumer(ctx context.Context, state *index.SegmentWriteState) (index.NormsConsumer, error) {
	return NewSimpleTextNormsConsumer(nil, state)
}

func (s *NormsFormat) NormsProducer(ctx context.Context, state *index.SegmentReadState) (index.NormsProducer, error) {
	return NewSimpleTextNormsProducer(state)
}

var _ index.NormsProducer = &SimpleTextNormsProducer{}

type SimpleTextNormsProducer struct {
	impl *DocValuesReader
}

func NewSimpleTextNormsProducer(state *index.SegmentReadState) (*SimpleTextNormsProducer, error) {
	reader, err := NewDocValuesReader(context.TODO(), state, NORMS_SEG_EXTENSION)
	if err != nil {
		return nil, err
	}
	return &SimpleTextNormsProducer{impl: reader}, nil
}

func (s *SimpleTextNormsProducer) GetNorms(field *document.FieldInfo) (index.NumericDocValues, error) {
	return s.impl.GetNumeric(nil, field)
}

func (s *SimpleTextNormsProducer) Close() error {
	return s.impl.Close()
}

func (s *SimpleTextNormsProducer) CheckIntegrity() error {
	return s.impl.CheckIntegrity()
}

func (s *SimpleTextNormsProducer) GetMergeInstance() index.NormsProducer {
	return s
}

var _ index.NormsConsumer = &SimpleTextNormsConsumer{}

// SimpleTextNormsConsumer
// Writes plain-text norms.
// FOR RECREATIONAL USE ONLY
type SimpleTextNormsConsumer struct {
	*coreIndex.NormsConsumerDefault

	dw *DocValuesWriter
}

func NewSimpleTextNormsConsumer(ctx context.Context, state *index.SegmentWriteState) (*SimpleTextNormsConsumer, error) {
	writer, err := NewDocValuesWriter(ctx, state, NORMS_SEG_EXTENSION)
	if err != nil {
		return nil, err
	}

	consumer := &SimpleTextNormsConsumer{
		dw: writer,
	}
	consumer.NormsConsumerDefault = &coreIndex.NormsConsumerDefault{
		FnAddNormsField: consumer.AddNormsField,
	}
	return consumer, nil
}

func (s *SimpleTextNormsConsumer) Close() error {
	return s.dw.Close()
}

func (s *SimpleTextNormsConsumer) AddNormsField(ctx context.Context, field *document.FieldInfo, normsProducer index.NormsProducer) error {
	producer := struct {
		*coreIndex.EmptyDocValuesProducer
	}{
		&coreIndex.EmptyDocValuesProducer{},
	}
	producer.FnGetNumeric = func(ctx context.Context, field *document.FieldInfo) (index.NumericDocValues, error) {
		return normsProducer.GetNorms(field)
	}

	return s.dw.AddNumericField(ctx, field, producer)
}
