package index

import (
	"context"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

type DocConsumer interface {
	ProcessDocument(ctx context.Context, docId int, document *document.Document) error

	Flush(ctx context.Context, state *index.SegmentWriteState) (*DocMap, error)

	Abort() error

	// GetHasDocValues Returns a DocIdSetIterator for the given field or null
	// if the field doesn't have doc values.
	GetHasDocValues(field string) types.DocIdSetIterator
}
