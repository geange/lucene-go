package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.StoredFieldsFormat = &SimpleTextStoredFieldsFormat{}

type SimpleTextStoredFieldsFormat struct {
}

func (s *SimpleTextStoredFieldsFormat) FieldsReader(directory store.Directory, si *index.SegmentInfo,
	fn *index.FieldInfos, context *store.IOContext) (index.StoredFieldsReader, error) {

	return NewSimpleTextStoredFieldsReader(directory, si, fn, context)
}

func (s *SimpleTextStoredFieldsFormat) FieldsWriter(directory store.Directory,
	si *index.SegmentInfo, context *store.IOContext) (index.StoredFieldsWriter, error) {

	return NewSimpleTextStoredFieldsWriter(directory, si.Name(), context)
}
