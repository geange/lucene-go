package spans

import (
	"github.com/geange/lucene-go/core/interface/index"
	index2 "github.com/geange/lucene-go/core/types"
)

// SpanCollector
// An interface defining the collection of postings information from the leaves of a Spans
// lucene.experimental
type SpanCollector interface {
	// CollectLeaf
	// Collect information from postings
	// postings: a PostingsEnum
	// position: – the position of the PostingsEnum
	// term: – the Term for this postings list
	CollectLeaf(postings index.PostingsEnum, position int, term *index2.Term) error

	// Reset
	// Call to indicate that the driving Spans has moved to a new position
	Reset()
}
