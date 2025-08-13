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
	return NewNormsConsumer(ctx, state, NORMS_DATA_CODEC, NORMS_DATA_EXTENSION, NORMS_METADATA_CODEC, NORMS_METADATA_EXTENSION)
}

func (n *NormsFormat) NormsProducer(ctx context.Context, state *index.SegmentReadState) (index.NormsProducer, error) {
	return NewNormsProducer(ctx, state, NORMS_DATA_CODEC, NORMS_DATA_EXTENSION, NORMS_METADATA_CODEC, NORMS_METADATA_EXTENSION)
}

const (
	NORMS_DATA_CODEC         = "Lucene80NormsData"
	NORMS_DATA_EXTENSION     = "nvd"
	NORMS_METADATA_CODEC     = "Lucene80NormsMetadata"
	NORMS_METADATA_EXTENSION = "nvm"
	NORMS_VERSION_START      = 0
	NORMS_VERSION_CURRENT    = NORMS_VERSION_START
)
