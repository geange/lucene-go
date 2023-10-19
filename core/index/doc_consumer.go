package index

import (
	"context"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/types"
)

type DocConsumer interface {
	ProcessDocument(docId int, document *document.Document) error

	Flush(ctx context.Context, state *SegmentWriteState) (*DocMap, error)

	Abort() error

	// GetHasDocValues Returns a DocIdSetIterator for the given field or null
	// if the field doesn't have doc values.
	GetHasDocValues(field string) types.DocIdSetIterator
}
