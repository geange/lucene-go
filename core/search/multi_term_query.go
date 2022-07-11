package search

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/tokenattributes"
)

// MultiTermQuery An abstract Query that matches documents containing a subset of terms provided by a
// FilteredTermsEnum enumeration.
// This query cannot be used directly; you must subclass it and define getTermsEnum(Terms, AttributeSource) to
// provide a FilteredTermsEnum that iterates through the terms to be matched.
// NOTE: if setRewriteMethod is either CONSTANT_SCORE_BOOLEAN_REWRITE or SCORING_BOOLEAN_REWRITE, you may
// encounter a BooleanQuery.TooManyClauses exception during searching, which happens when the number of terms
// to be searched exceeds BooleanQuery.getMaxClauseCount(). Setting setRewriteMethod to CONSTANT_SCORE_REWRITE
// prevents this.
// The recommended rewrite method is CONSTANT_SCORE_REWRITE: it doesn't spend CPU computing unhelpful scores,
// and is the most performant rewrite method given the query. If you need scoring (like FuzzyQuery,
// use MultiTermQuery.TopTermsScoringBooleanQueryRewrite which uses a priority queue to only collect
// competitive terms and not hit this limitation. Note that org.apache.lucene.queryparser.classic.QueryParser
// produces MultiTermQueries using CONSTANT_SCORE_REWRITE by default.
type MultiTermQuery interface {
	Query
}

// RewriteMethod Abstract class that defines how the query is rewritten.
type RewriteMethod interface {
	Rewrite(reader index.IndexReader, query MultiTermQuery) (Query, error)

	// GetTermsEnum Returns the MultiTermQuerys TermsEnum
	// See Also: getTermsEnum(Terms, AttributeSource)
	GetTermsEnum(query MultiTermQuery, terms index.Terms, atts *tokenattributes.AttributeSource) (index.TermsEnum, error)
}
