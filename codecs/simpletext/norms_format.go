package simpletext

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
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

func (s *NormsFormat) NormsConsumer(state *index.SegmentWriteState) (index.NormsConsumer, error) {
	return NewSimpleTextNormsConsumer(state)
}

func (s *NormsFormat) NormsProducer(state *index.SegmentReadState) (index.NormsProducer, error) {
	return NewSimpleTextNormsProducer(state)
}

var _ index.NormsProducer = &SimpleTextNormsProducer{}

type SimpleTextNormsProducer struct {
	impl *DocValuesReader
}

func NewSimpleTextNormsProducer(state *index.SegmentReadState) (*SimpleTextNormsProducer, error) {
	reader, err := NewDocValuesReader(state, NORMS_SEG_EXTENSION)
	if err != nil {
		return nil, err
	}
	return &SimpleTextNormsProducer{impl: reader}, nil
}

func (s *SimpleTextNormsProducer) GetNorms(field *document.FieldInfo) (index.NumericDocValues, error) {
	return s.impl.GetNumeric(field)
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

// SimpleTextNormsConsumer Writes plain-text norms.
// FOR RECREATIONAL USE ONLY
type SimpleTextNormsConsumer struct {
	*index.NormsConsumerDefault

	impl *DocValuesWriter
}

func NewSimpleTextNormsConsumer(state *index.SegmentWriteState) (*SimpleTextNormsConsumer, error) {
	writer, err := NewDocValuesWriter(state, NORMS_SEG_EXTENSION)
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

func (s *SimpleTextNormsConsumer) AddNormsField(field *document.FieldInfo, normsProducer index.NormsProducer) error {
	producer := struct {
		*index.EmptyDocValuesProducer
	}{}
	producer.FnGetNumeric = func(field *document.FieldInfo) (index.NumericDocValues, error) {
		return normsProducer.GetNorms(field)
	}

	return s.impl.AddNumericField(nil, field, producer)
}
