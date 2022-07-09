package search

import "github.com/geange/lucene-go/core/index"

// Weight Expert: Calculate query weights and build query scorers.
// The purpose of Weight is to ensure searching does not modify a Query, so that a Query instance can be reused.
// IndexSearcher dependent state of the query should reside in the Weight. LeafReader dependent state should
// reside in the Scorer.
// Since Weight creates Scorer instances for a given LeafReaderContext (scorer(LeafReaderContext)) callers must
// maintain the relationship between the searcher's top-level IndexReaderContext and the context used to create
// a Scorer.
// A Weight is used in the following way:
// A Weight is constructed by a top-level query, given a IndexSearcher (Query.createWeight(IndexSearcher, ScoreMode, float)).
// A Scorer is constructed by scorer(LeafReaderContext).
// Since: 2.9
type Weight interface {
	// Matches Returns Matches for a specific document, or null if the document does not match the parent query A query match that contains no position information (for example, a Point or DocValues query) will return MatchesUtils.MATCH_WITH_NO_TERMS
	// Params: 	context – the reader's context to create the Matches for
	//			doc – the document's id relative to the given context's reader
	Matches(context *index.LeafReaderContext, doc int) (Matches, error)

	Match(value interface{}, description string, details []Explanation) (*Explanation, error)

	NoMatch(value interface{}, description string, details []Explanation) (*Explanation, error)

	// IsMatch Indicates whether or not this Explanation models a match.
	IsMatch() bool

	// GetValue The value assigned to this explanation node.
	GetValue() any

	// GetDescription A description of this explanation node.
	GetDescription() string

	//GetSummary() string

	// GetDetails The sub-nodes of this explanation node.
	GetDetails() []Explanation
}
