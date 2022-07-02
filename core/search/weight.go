package search

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
}
