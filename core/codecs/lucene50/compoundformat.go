package lucene50

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.CompoundFormat = &CompoundFormat{}

type CompoundFormat struct {
}

func (f *CompoundFormat) GetCompoundReader(CompoundFormat context.Context, dir store.Directory, si index.SegmentInfo, context *store.IOContext) (index.CompoundDirectory, error) {
	//TODO implement me
	panic("implement me")
}

func (f *CompoundFormat) Write(ctx context.Context, dir store.Directory, si index.SegmentInfo, ioContext *store.IOContext) error {
	//TODO implement me
	panic("implement me")
}
