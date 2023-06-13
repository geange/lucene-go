package spans

import "github.com/geange/lucene-go/core/index"

// SpanCollector
// An interface defining the collection of postings information from the leaves of a Spans
// lucene.experimental
type SpanCollector interface {
	// CollectLeaf Collect information from postings
	// Params: 	postings – a PostingsEnum
	//			position – the position of the PostingsEnum
	//			term – the Term for this postings list
	// Throws: IOException – on error
	CollectLeaf(postings index.PostingsEnum, position int, term *index.Term) error

	// Reset
	// Call to indicate that the driving Spans has moved to a new position
	Reset()
}
