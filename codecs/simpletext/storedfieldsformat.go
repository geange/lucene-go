package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.StoredFieldsFormat = &StoredFieldsFormat{}

type StoredFieldsFormat struct {
}

func NewStoredFieldsFormat() *StoredFieldsFormat {
	return &StoredFieldsFormat{}
}

func (s *StoredFieldsFormat) FieldsReader(directory store.Directory, si *index.SegmentInfo,
	fn *index.FieldInfos, context *store.IOContext) (index.StoredFieldsReader, error) {

	return NewStoredFieldsReader(directory, si, fn, context)
}

func (s *StoredFieldsFormat) FieldsWriter(directory store.Directory,
	si *index.SegmentInfo, context *store.IOContext) (index.StoredFieldsWriter, error) {

	return NewStoredFieldsWriter(directory, si.Name(), context)
}
