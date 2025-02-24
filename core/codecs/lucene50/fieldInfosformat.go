package lucene50

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.FieldInfosFormat = &FieldInfosFormat{}

type FieldInfosFormat struct {
}

func (f *FieldInfosFormat) Read(ctx context.Context, directory store.Directory, segmentInfo index.SegmentInfo, segmentSuffix string, ioContext *store.IOContext) (index.FieldInfos, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FieldInfosFormat) Write(ctx context.Context, directory store.Directory, segmentInfo index.SegmentInfo, segmentSuffix string, infos index.FieldInfos, ioContext *store.IOContext) error {
	//TODO implement me
	panic("implement me")
}
