package search

// MatchesIterator An iterator over match positions (and optionally offsets) for a single document and field To iterate over the matches, call next() until it returns false, retrieving positions and/or offsets after each call. You should not call the position or offset methods before next() has been called, or after next() has returned false. Matches from some queries may span multiple positions. You can retrieve the positions of individual matching terms on the current match by calling getSubMatches(). Matches are ordered by start position, and then by end position. Match intervals may overlap.
// See Also: Weight.matches(LeafReaderContext, int)
type MatchesIterator interface {
}
