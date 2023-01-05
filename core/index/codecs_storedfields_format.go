package index

import "github.com/geange/lucene-go/core/store"

type StoredFieldsFormat interface {

	// FieldsReader Returns a StoredFieldsReader to load stored fields.
	FieldsReader(directory store.Directory, si *SegmentInfo,
		fn *FieldInfos, context *store.IOContext) (StoredFieldsReader, error)

	// FieldsWriter Returns a StoredFieldsWriter to write stored fields.
	FieldsWriter(directory store.Directory,
		si *SegmentInfo, context *store.IOContext) (StoredFieldsWriter, error)
}
