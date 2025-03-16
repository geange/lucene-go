package compressing

import (
	"context"
	"errors"
	"fmt"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed"
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

func NewStoredFieldsFormat(formatName, segmentSuffix string, compressionMode CompressionMode,
	chunkSize, maxDocsPerChunk, blockShift int) (*StoredFieldsFormat, error) {

	this := &StoredFieldsFormat{
		formatName:      formatName,
		segmentSuffix:   segmentSuffix,
		compressionMode: compressionMode,
	}
	if chunkSize < 1 {
		return nil, errors.New("chunkSize must be >= 1")
	}
	this.chunkSize = chunkSize
	if maxDocsPerChunk < 1 {
		return nil, errors.New("maxDocsPerChunk must be >= 1")
	}
	this.maxDocsPerChunk = maxDocsPerChunk
	if blockShift < packed.MIN_BLOCK_SHIFT || blockShift > packed.MAX_BLOCK_SHIFT {
		return nil, fmt.Errorf("blockSize must be in %d-%d, get %d", packed.MIN_BLOCK_SHIFT, packed.MAX_BLOCK_SHIFT, blockShift)
	}
	this.blockShift = blockShift

	return this, nil
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
