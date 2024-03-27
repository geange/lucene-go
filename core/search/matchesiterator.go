package search

// MatchesIterator An iterator over match positions (and optionally offsets) for a single document and field To iterate over the matches, call next() until it returns false, retrieving positions and/or offsets after each call. You should not call the position or offset methods before next() has been called, or after next() has returned false. Matches from some queries may span multiple positions. You can retrieve the positions of individual matching terms on the current match by calling getSubMatches(). Matches are ordered by start position, and then by end position. Match intervals may overlap.
// See Also: Weight.matches(LeafReaderContext, int)
type MatchesIterator interface {

	// Next Advance the iterator to the next match position
	// Returns: true if matches have not been exhausted
	Next() (bool, error)

	// StartPosition The start position of the current match OccurShould only be called after next() has returned true
	StartPosition() int

	// EndPosition The end position of the current match OccurShould only be called after next() has returned true
	EndPosition() int

	// StartOffset The starting offset of the current match, or -1 if offsets are not available OccurShould only be
	// called after next() has returned true
	StartOffset() (int, error)

	// EndOffset The ending offset of the current match, or -1 if offsets are not available OccurShould only be
	// called after next() has returned true
	EndOffset() (int, error)

	// GetSubMatches Returns a MatchesIterator that iterates over the positions and offsets of individual
	// terms within the current match Returns null if there are no submatches (ie the current iterator is
	// at the leaf level) OccurShould only be called after next() has returned true
	GetSubMatches() (MatchesIterator, error)

	// GetQuery Returns the Query causing the current match If this MatchesIterator has been returned from a
	// getSubMatches() call, then returns a TermQuery equivalent to the current match OccurShould only be called
	// after next() has returned true
	GetQuery() Query
}
