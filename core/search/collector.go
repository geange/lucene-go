package search

import "github.com/geange/lucene-go/core/index"

// Collector Expert: Collectors are primarily meant to be used to gather raw results from a search, and implement sorting or custom result filtering, collation, etc.
// Lucene's core collectors are derived from Collector and SimpleCollector. Likely your application can use one of these classes, or subclass TopDocsCollector, instead of implementing Collector directly:
// TopDocsCollector is an abstract base class that assumes you will retrieve the top N docs, according to some criteria, after collection is done.
// TopScoreDocCollector is a concrete subclass TopDocsCollector and sorts according to Score + docID. This is used internally by the IndexSearcher search methods that do not take an explicit Sort. It is likely the most frequently used collector.
// TopFieldCollector subclasses TopDocsCollector and sorts according to a specified Sort object (sort by field). This is used internally by the IndexSearcher search methods that take an explicit Sort.
// TimeLimitingCollector, which wraps any other Collector and aborts the search if it's taken too much time.
// PositiveScoresOnlyCollector wraps any other Collector and prevents collection of hits whose Score is <= 0.0
type Collector interface {

	// GetLeafCollector Create a new collector to collect the given context.
	// Params: context – next atomic reader context
	GetLeafCollector(context *index.LeafReaderContext) (LeafCollector, error)

	// ScoreMode Indicates what features are required from the scorer.
	ScoreMode() *ScoreMode
}
