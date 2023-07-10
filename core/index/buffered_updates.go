package index

import (
	"github.com/geange/gods-generic/maps/treemap"
	"go.uber.org/atomic"
)

// BufferedUpdates Holds buffered deletes and updates, by docID, term or query for a single segment.
// This is used to hold buffered pending deletes and updates against the to-be-flushed segment.
// Once the deletes and updates are pushed (on Flush in DocumentsWriter), they are converted to a
// FrozenBufferedUpdates instance and pushed to the BufferedUpdatesStream.
//
// NOTE: instances of this class are accessed either via a private
// instance on DocumentWriterPerThread, or via sync'd code by
// DocumentsWriterDeleteQueue
type BufferedUpdates struct {
	numTermDeletes  *atomic.Int64
	numFieldUpdates *atomic.Int64
	deleteTerms     *treemap.Map[*Term, int]
	//deleteQueries   *treemap.Map
	gen         int64
	segmentName string
}

func NewBufferedUpdates() *BufferedUpdates {
	return &BufferedUpdates{
		numTermDeletes:  nil,
		numFieldUpdates: nil,
		deleteTerms:     treemap.NewWith[*Term, int](TermCompare),
	}
}

func NewBufferedUpdatesV1(segmentName string) *BufferedUpdates {
	return &BufferedUpdates{
		numTermDeletes:  nil,
		numFieldUpdates: nil,
		deleteTerms:     treemap.NewWith[*Term, int](TermCompare),
		segmentName:     segmentName,
	}
}

func (b *BufferedUpdates) AddTerm(term *Term, docIDUpto int) {
	current, ok := b.deleteTerms.Get(term)
	if ok {
		if current > docIDUpto {
			// Only record the new number if it's greater than the
			// current one.  This is important because if multiple
			// threads are replacing the same doc at nearly the
			// same time, it's possible that one thread that got a
			// higher docID is scheduled before the other
			// threads.  If we blindly replace than we can
			// incorrectly get both docs indexed.
			return
		}
	}

	b.deleteTerms.Put(term, docIDUpto)
	// note that if current != null then it means there's already a buffered
	// delete on that term, therefore we seem to over-count. this over-counting
	// is done to respect IndexWriterConfig.setMaxBufferedDeleteTerms.
	b.numTermDeletes.Inc()
}
