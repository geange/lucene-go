package index

import (
	"github.com/geange/lucene-go/core/document"
)

type DocConsumer interface {
	ProcessDocument(docId int, document *document.Document) error

	Flush(state *SegmentWriteState) (*DocMap, error)

	Abort() error

	// GetHasDocValues Returns a DocIdSetIterator for the given field or null
	// if the field doesn't have doc values.
	GetHasDocValues(field string) DocIdSetIterator
}
