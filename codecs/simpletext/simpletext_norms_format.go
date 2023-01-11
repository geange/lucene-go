package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.NormsFormat = &SimpleTextNormsFormat{}

const (
	NORMS_SEG_EXTENSION = "len"
)

// SimpleTextNormsFormat plain-text norms format.
// FOR RECREATIONAL USE ONLY
type SimpleTextNormsFormat struct {
}

func (s *SimpleTextNormsFormat) NormsConsumer(state *index.SegmentWriteState) (index.NormsConsumer, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextNormsFormat) NormsProducer(state *index.SegmentReadState) (index.NormsProducer, error) {
	//TODO implement me
	panic("implement me")
}

var _ index.NormsProducer = &SimpleTextNormsProducer{}

type SimpleTextNormsProducer struct {
	impl *SimpleTextDocValuesReader
}

func NewSimpleTextNormsProducer(state *index.SegmentReadState) (*SimpleTextNormsProducer, error) {
	reader, err := NewSimpleTextDocValuesReader(state, NORMS_SEG_EXTENSION)
	if err != nil {
		return nil, err
	}
	return &SimpleTextNormsProducer{impl: reader}, nil
}

func (s *SimpleTextNormsProducer) GetNorms(field *types.FieldInfo) (index.NumericDocValues, error) {
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

	impl *SimpleTextDocValuesWriter
}

func (s *SimpleTextNormsConsumer) Close() error {
	return s.impl.Close()
}

func (s *SimpleTextNormsConsumer) AddNormsField(field *types.FieldInfo, normsProducer index.NormsProducer) error {
	producer := struct {
		*index.EmptyDocValuesProducer
	}{}
	producer.FnGetNumeric = func(field *types.FieldInfo) (index.NumericDocValues, error) {
		return normsProducer.GetNorms(field)
	}

	return s.impl.AddNumericField(field, producer)
}
