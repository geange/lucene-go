package index

import "github.com/geange/lucene-go/core/types"

type DocConsumer interface {
	ProcessDocument(docId int, document []types.IndexableField) error
	Flush(state *SegmentWriteState) (DocMap, error)
	Abort() error

	// GetHasDocValues Returns a DocIdSetIterator for the given field or null if the field doesn't have doc values.
	GetHasDocValues(field string) DocIdSetIterator
}
