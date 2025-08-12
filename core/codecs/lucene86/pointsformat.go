package lucene86

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
)

const (
	POINT_DATA_CODEC_NAME  = "Lucene86PointsFormatData"
	POINT_INDEX_CODEC_NAME = "Lucene86PointsFormatIndex"
	POINT_META_CODEC_NAME  = "Lucene86PointsFormatMeta"
	POINT_DATA_EXTENSION   = "kdd" // Filename extension for the leaf blocks
	POINT_INDEX_EXTENSION  = "kdi" // Filename extension for the index per field
	POINT_META_EXTENSION   = "kdm" // Filename extension for the meta per field

	POINT_VERSION_START   = 0
	POINT_VERSION_CURRENT = POINT_VERSION_START
)

var _ index.PointsFormat = &PointsFormat{}

type PointsFormat struct {
}

func NewPointsFormat() *PointsFormat {
	return &PointsFormat{}
}

func (p *PointsFormat) FieldsWriter(ctx context.Context, state *index.SegmentWriteState) (index.PointsWriter, error) {
	return NewPointsWriter(ctx, NewPointsWriterConfig(state))
}

func (p *PointsFormat) FieldsReader(ctx context.Context, state *index.SegmentReadState) (index.PointsReader, error) {
	return NewPointsReader(ctx, state)
}
