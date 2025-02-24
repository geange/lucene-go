package lucene50

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
)

var _ index.LiveDocsFormat = &LiveDocsFormat{}

type LiveDocsFormat struct {
}

func (f *LiveDocsFormat) ReadLiveDocs(ctx context.Context, directory store.Directory, info index.SegmentCommitInfo, context *store.IOContext) (util.Bits, error) {
	//TODO implement me
	panic("implement me")
}

func (f *LiveDocsFormat) WriteLiveDocs(ctx context.Context, bits util.Bits, directory store.Directory, info index.SegmentCommitInfo, newDelCount int, ioContext *store.IOContext) error {
	//TODO implement me
	panic("implement me")
}

func (f *LiveDocsFormat) Files(ctx context.Context, info index.SegmentCommitInfo, files map[string]struct{}) (map[string]struct{}, error) {
	//TODO implement me
	panic("implement me")
}
