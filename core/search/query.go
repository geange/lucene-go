package search

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
}
