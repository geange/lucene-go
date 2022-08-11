package index

import (
	"github.com/emirpasic/gods/maps/treemap"
	"go.uber.org/atomic"
)

// BufferedUpdates Holds buffered deletes and updates, by docID, term or query for a single segment.
// This is used to hold buffered pending deletes and updates against the to-be-flushed segment.
// Once the deletes and updates are pushed (on flush in DocumentsWriter), they are converted to a
// FrozenBufferedUpdates instance and pushed to the BufferedUpdatesStream.
//
// NOTE: instances of this class are accessed either via a private
// instance on DocumentWriterPerThread, or via sync'd code by
// DocumentsWriterDeleteQueue
type BufferedUpdates struct {
	numTermDeletes  *atomic.Int64
	numFieldUpdates *atomic.Int64
	deleteTerms     *treemap.Map
	deleteQueries   *treemap.Map
}

func NewBufferedUpdates() *BufferedUpdates {
	return &BufferedUpdates{
		numTermDeletes:  nil,
		numFieldUpdates: nil,
		deleteTerms:     treemap.NewWith(TermComparator),
	}
}
