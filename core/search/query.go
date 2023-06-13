package search

import "github.com/geange/lucene-go/core/index"

// Query The abstract base class for queries.
// * Instantiable subclasses are:
// * TermQuery
// * BooleanQuery
// * WildcardQuery
// * PhraseQuery
// * PrefixQuery
// * MultiPhraseQuery
// * FuzzyQuery
// * RegexpQuery
// * TermRangeQuery
// * PointRangeQuery
// * ConstantScoreQuery
// * DisjunctionMaxQuery
// * MatchAllDocsQuery
// See also the family of Span Queries and additional queries available in the Queries module
type Query interface {

	// String
	// ToString Prints a query to a string, with field assumed to be the default field and omitted.
	// ToString(field string) string
	String(field string) string

	// CreateWeight
	// Expert: Constructs an appropriate Weight implementation for this query.
	// Only implemented by primitive queries, which re-write to themselves.
	// Params: 	scoreMode – How the produced scorers will be consumed.
	//			boost – The boost that is propagated by the parent queries.
	CreateWeight(searcher *IndexSearcher, scoreMode *ScoreMode, boost float64) (Weight, error)

	// Rewrite
	// Expert: called to re-write queries into primitive queries. For example, a PrefixQuery will be
	// rewritten into a BooleanQuery that consists of TermQuerys.
	Rewrite(reader index.IndexReader) (Query, error)

	// Visit
	// Recurse through the query tree, visiting any child queries
	// Params: visitor – a QueryVisitor to be called by each query in the tree
	Visit(visitor QueryVisitor) (err error)
}
