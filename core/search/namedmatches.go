package search

import (
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.Matches = &NamedMatches{}

// NamedMatches
// Utility class to help extract the set of sub queries that have matched from a larger query.
// Individual subqueries may be wrapped using wrapQuery(String, Query), and the matching queries for a
// particular document can then be pulled from the parent Query's Matches object by calling findNamedMatches(Matches)
type NamedMatches struct {
	in   index.Matches
	name string
}

func NewNamedMatches(in index.Matches, name string) *NamedMatches {
	return &NamedMatches{in: in, name: name}
}

func (n *NamedMatches) GetName() string {
	return n.name
}

func (n *NamedMatches) Strings() []string {
	return n.in.Strings()
}

func (n *NamedMatches) GetMatches(field string) (index.MatchesIterator, error) {
	return n.in.GetMatches(field)
}

func (n *NamedMatches) GetSubMatches() []index.Matches {
	return []index.Matches{n.in}
}
