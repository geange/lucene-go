package index

import "go.uber.org/atomic"

// ReadersAndUpdates
// Used by IndexWriter to hold open SegmentReaders (for
// searching or merging), plus pending deletes and updates,
// for a given segment
type ReadersAndUpdates struct {
	// Not final because we replace (clone) when we need to
	// change it and it's been shared:
	info *SegmentCommitInfo

	// Tracks how many consumers are using this instance:
	refCount *atomic.Int64

	// Set once (null, and then maybe set, and never set again):
	reader *SegmentReader

	// How many further deletions we've done against
	// liveDocs vs when we loaded it or last wrote it:
	pendingDeletes PendingDeletes

	// the major version this index was created with
	indexCreatedVersionMajor int

	// Indicates whether this segment is currently being merged. While a segment
	// is merging, all field updates are also registered in the
	// mergingNumericUpdates map. Also, calls to writeFieldUpdates merge the
	// updates with mergingNumericUpdates.
	// That way, when the segment is done merging, IndexWriter can apply the
	// updates on the merged segment too.
	isMerging bool

	// Holds resolved (to docIDs) doc values updates that have not yet been
	// written to the index
	pendingDVUpdates map[string][]DocValuesFieldUpdates
}

func NewReadersAndUpdatesV1(indexCreatedVersionMajor int,
	reader *SegmentReader, pendingDeletes PendingDeletes) (*ReadersAndUpdates, error) {

	panic("")
}
