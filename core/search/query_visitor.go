package search

import "github.com/geange/lucene-go/core/index"

// QueryVisitor
// Allows recursion through a query tree
// See Also: Query.visit(QueryVisitor)
type QueryVisitor interface {

	// ConsumeTerms
	// Called by leaf queries that match on specific terms
	// Params:
	//			query – the leaf query
	//			terms – the terms the query will match on
	ConsumeTerms(query Query, terms ...*index.Term)

	// Called by leaf queries that match on a class of terms
	// Params:
	//			query – the leaf query
	//			field – the field queried against
	//			automaton – a supplier for an automaton defining which terms match
	//consumeTermsMatching(query Query, )

	// VisitLeaf Called by leaf queries that do not match on terms
	// Params: query – the query
	VisitLeaf(query Query) (err error)

	// AcceptField
	// Whether or not terms from this field are of interest to the visitor Implement this to
	// avoid collecting terms from heavy queries such as TermInSetQuery that are not running
	// on fields of interest
	AcceptField(field string) bool

	// GetSubVisitor
	// Pulls a visitor instance for visiting child clauses of a query The default implementation
	// returns this, unless occur is equal to BooleanClause.Occur.OccurMustNot in which case it
	// returns EMPTY_VISITOR
	// Params:
	//			occur – the relationship between the parent and its children
	//			parent – the query visited
	GetSubVisitor(occur Occur, parent Query) QueryVisitor
}
