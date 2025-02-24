package compressing

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.TermVectorsFormat = &TermVectorsFormat{}

type TermVectorsFormat struct {
	formatName      string
	segmentSuffix   string
	compressionMode CompressionMode
	chunkSize       int
	blockSize       int
	maxDocsPerChunk int
}

func NewTermVectorsFormat(formatName, segmentSuffix string, compressionMode CompressionMode,
	chunkSize, maxDocsPerChunk, blockSize int) *TermVectorsFormat {
	return &TermVectorsFormat{formatName: formatName, segmentSuffix: segmentSuffix, compressionMode: compressionMode, chunkSize: chunkSize, blockSize: blockSize, maxDocsPerChunk: maxDocsPerChunk}
}

func (f *TermVectorsFormat) VectorsReader(ctx context.Context, directory store.Directory, segmentInfo index.SegmentInfo, fieldInfos index.FieldInfos, ioContext *store.IOContext) (index.TermVectorsReader, error) {
	//TODO implement me
	panic("implement me")
}

func (f *TermVectorsFormat) VectorsWriter(ctx context.Context, directory store.Directory, segmentInfo index.SegmentInfo, ioContext *store.IOContext) (index.TermVectorsWriter, error) {
	//TODO implement me
	panic("implement me")
}
