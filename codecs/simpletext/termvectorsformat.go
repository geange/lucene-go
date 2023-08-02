package simpletext

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.TermVectorsFormat = &TermVectorsFormat{}

type TermVectorsFormat struct {
}

func NewTermVectorsFormat() *TermVectorsFormat {
	return &TermVectorsFormat{}
}

func (s *TermVectorsFormat) VectorsReader(ctx context.Context, dir store.Directory, segmentInfo index.SegmentInfo, fieldInfos index.FieldInfos, ioContext *store.IOContext) (index.TermVectorsReader, error) {
	return NewTermVectorsReader(ctx, dir, segmentInfo, ioContext)
}

func (s *TermVectorsFormat) VectorsWriter(ctx context.Context, dir store.Directory, segmentInfo index.SegmentInfo, ioContext *store.IOContext) (index.TermVectorsWriter, error) {
	return NewTermVectorsWriter(ctx, dir, segmentInfo.Name(), ioContext)
}
