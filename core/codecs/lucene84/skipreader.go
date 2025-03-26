package lucene84

import (
	"context"

	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.MultiLevelSkipListReaderSPI = &SkipReader{}

type SkipReader struct {
	docPointer          []int64
	posPointer          []int64
	payPointer          []int64
	posBufferUpto       []int
	payloadByteUpto     []int
	lastPosPointer      int64
	lastPayPointer      int64
	lastPayloadByteUpto int
	lastDocPointer      int64
	lastPosBufferUpto   int
}

func (s *SkipReader) ReadSkipData(ctx context.Context, level int, skipStream store.IndexInput, mtx *index.MultiLevelSkipListReaderContext) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SkipReader) ReadLevelLength(ctx context.Context, skipStream store.IndexInput, mtx *index.MultiLevelSkipListReaderContext) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SkipReader) ReadChildPointer(ctx context.Context, skipStream store.IndexInput, mtx *index.MultiLevelSkipListReaderContext) (int64, error) {
	//TODO implement me
	panic("implement me")
}
