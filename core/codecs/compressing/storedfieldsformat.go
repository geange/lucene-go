package compressing

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.StoredFieldsFormat = &StoredFieldsFormat{}

type StoredFieldsFormat struct {
	formatName      string
	segmentSuffix   string
	compressionMode CompressionMode
	chunkSize       int
	maxDocsPerChunk int
	blockShift      int
}

func (s *StoredFieldsFormat) FieldsReader(ctx context.Context, directory store.Directory, si index.SegmentInfo,
	fn index.FieldInfos, ioContext *store.IOContext) (index.StoredFieldsReader, error) {

	return NewStoredFieldsReader(ctx, directory, si, s.segmentSuffix, fn, ioContext, s.formatName, s.compressionMode)
}

func (s *StoredFieldsFormat) FieldsWriter(ctx context.Context, directory store.Directory, si index.SegmentInfo,
	ioContext *store.IOContext) (index.StoredFieldsWriter, error) {

	return NewStoredFieldsWriter(ctx, directory, si, s.segmentSuffix, ioContext,
		s.formatName, s.compressionMode, s.chunkSize, s.maxDocsPerChunk, s.blockShift)
}
