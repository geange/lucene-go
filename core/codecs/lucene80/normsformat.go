package lucene80

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.NormsFormat = &NormsFormat{}

type NormsFormat struct {
}

func NewNormsFormat() *NormsFormat {
	return &NormsFormat{}
}

func (n *NormsFormat) NormsConsumer(ctx context.Context, state *index.SegmentWriteState) (index.NormsConsumer, error) {
	//TODO implement me
	panic("implement me")
}

func (n *NormsFormat) NormsProducer(ctx context.Context, state *index.SegmentReadState) (index.NormsProducer, error) {
	//TODO implement me
	panic("implement me")
}
