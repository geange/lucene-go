package search

// Matches Reports the positions and optionally offsets of all matching terms in a query for a single document
// To obtain a MatchesIterator for a particular field, call GetMatches(String). Note that you can call
// GetMatches(String) multiple times to retrieve new iterators, but it is not thread-safe.
type Matches interface {
	Strings() []string

	// GetMatches Returns a MatchesIterator over the matches for a single field, or null if there are no matches
	// in that field.
	GetMatches(field string) (MatchesIterator, error)

	// GetSubMatches Returns a collection of Matches that make up this instance; if it is not a composite,
	// then this returns an empty list
	GetSubMatches() []Matches
}

type MatchesImp struct {
	strs []string
}
