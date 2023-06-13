package spans

import "github.com/geange/lucene-go/core/index"

// Spans
// Iterates through combinations of start/end positions per-doc. Each start/end position represents a range of term positions within the current document. These are enumerated in order, by increasing document number, within that by increasing start position and finally by increasing end position.
type Spans interface {
	index.DocIdSetIterator

	// NextStartPosition
	// Returns the next start position for the current doc. There is always at least one start/end position per doc. After the last start/end position at the current doc this returns NO_MORE_POSITIONS.
	NextStartPosition() (int, error)

	// StartPosition
	// Returns the start position in the current doc, or -1 when nextStartPosition was not yet called on the current doc. After the last start/end position at the current doc this returns NO_MORE_POSITIONS.
	StartPosition() int

	// EndPosition
	// Returns the end position for the current start position, or -1 when nextStartPosition was not yet called on the current doc. After the last start/end position at the current doc this returns NO_MORE_POSITIONS.
	EndPosition() int

	// Width
	// Return the width of the match, which is typically used to sloppy freq. It is only legal to call this method when the iterator is on a valid doc ID and positioned. The return value must be positive, and lower values means that the match is better.
	Width() int

	// Collect postings data from the leaves of the current Spans. This method should only be called after nextStartPosition(), and before NO_MORE_POSITIONS has been reached.
	// Params: collector – a SpanCollector
	// lucene.experimental
	collect(collector SpanCollector)
}
