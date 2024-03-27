package simpletext

import (
	"context"

	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.StoredFieldsFormat = &StoredFieldsFormat{}

type StoredFieldsFormat struct {
}

func NewStoredFieldsFormat() *StoredFieldsFormat {
	return &StoredFieldsFormat{}
}

func (s *StoredFieldsFormat) FieldsReader(ctx context.Context, directory store.Directory, si *index.SegmentInfo, fn *index.FieldInfos, ioContext *store.IOContext) (index.StoredFieldsReader, error) {

	return NewStoredFieldsReader(ctx, directory, si, fn, ioContext)
}

func (s *StoredFieldsFormat) FieldsWriter(ctx context.Context, directory store.Directory, si *index.SegmentInfo, ioContext *store.IOContext) (index.StoredFieldsWriter, error) {
	return NewStoredFieldsWriter(ctx, directory, si.Name(), ioContext)
}
