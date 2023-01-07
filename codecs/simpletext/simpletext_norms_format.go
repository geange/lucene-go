package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.NormsFormat = &SimpleTextNormsFormat{}

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
}

func (s *SimpleTextNormsProducer) GetNorms(field *types.FieldInfo) (index.NumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextNormsProducer) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextNormsProducer) GetMergeInstance() index.NormsProducer {
	//TODO implement me
	panic("implement me")
}

var _ index.NormsConsumer = &SimpleTextNormsConsumer{}

// SimpleTextNormsConsumer Writes plain-text norms.
// FOR RECREATIONAL USE ONLY
type SimpleTextNormsConsumer struct {
	*index.NormsConsumerImp
}

func (s *SimpleTextNormsConsumer) Merge(mergeState *index.MergeState) error {
	panic("")
}
