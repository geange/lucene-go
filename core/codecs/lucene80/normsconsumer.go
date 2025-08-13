package lucene80

import (
	"context"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.NormsConsumer = &NormsConsumer{}

type NormsConsumer struct {
}

func NewNormsConsumer(ctx context.Context, state *index.SegmentWriteState,
	dataCodec, dataExtension, metaCodec, metaExtension string) (*NormsConsumer, error) {
	panic("")
}

func (n *NormsConsumer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (n *NormsConsumer) AddNormsField(ctx context.Context, field *document.FieldInfo, normsProducer index.NormsProducer) error {
	//TODO implement me
	panic("implement me")
}

func (n *NormsConsumer) Merge(ctx context.Context, mergeState *index.MergeState) error {
	//TODO implement me
	panic("implement me")
}

func (n *NormsConsumer) MergeNormsField(ctx context.Context, mergeFieldInfo *document.FieldInfo, mergeState *index.MergeState) error {
	//TODO implement me
	panic("implement me")
}
