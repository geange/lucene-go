package index

import (
	"context"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/types"
)

// DocMap
// A permutation of doc IDs. For every document ID between 0 and Reader.maxDoc(),
// oldToNew(newToOld(docID)) must return docID.
type DocMap interface {

	// OldToNew
	// Given a doc ID from the original index, return its ordinal in the sorted index.
	OldToNew(docID int) int

	// NewToOld
	// Given the ordinal of a doc ID, return its doc ID in the original index.
	NewToOld(docID int) int

	// Size
	// Return the number of documents in this map.
	// This must be equal to the number of documents of the LeafReader which is sorted.
	Size() int
}

type DocConsumer interface {
	ProcessDocument(ctx context.Context, docId int, document *document.Document) error

	Flush(ctx context.Context, state *SegmentWriteState) (DocMap, error)

	Abort() error

	// GetHasDocValues Returns a DocIdSetIterator for the given field or null
	// if the field doesn't have doc values.
	GetHasDocValues(field string) types.DocIdSetIterator
}
