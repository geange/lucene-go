package search

import "fmt"

// BooleanClause
// A clause in a BooleanQuery.
type BooleanClause struct {
	// The query whose matching documents are combined by the boolean query.
	query Query

	occur Occur
}

func NewBooleanClause(query Query, occur Occur) *BooleanClause {
	return &BooleanClause{occur: occur, query: query}
}

func (b *BooleanClause) GetOccur() Occur {
	return b.occur
}

func (b *BooleanClause) GetQuery() Query {
	return b.query
}

func (b *BooleanClause) IsProhibited() bool {
	return MUST_NOT == b.occur
}

func (b *BooleanClause) IsRequired() bool {
	return b.occur == MUST || b.occur == FILTER
}

func (b *BooleanClause) IsScoring() bool {
	return b.occur == MUST || b.occur == SHOULD
}

func (b *BooleanClause) String() string {
	return fmt.Sprintf("%s %s", b.occur, b.query)
}

// Occur
// Specifies how clauses are to occur in matching documents.
type Occur string

func (o Occur) String() string {
	return string(o)
}

const (
	// MUST
	// Use this operator for clauses that must appear in the matching documents.
	MUST = Occur("+")

	// FILTER
	// Like MUST except that these clauses do not participate in scoring.
	FILTER = Occur("#")

	// SHOULD
	// Use this operator for clauses that should appear in the matching documents.
	// For a BooleanQuery with no MUST clauses one or more SHOULD clauses must match
	// a document for the BooleanQuery to match.
	// See Also: BooleanQuery.BooleanQueryBuilder.setMinimumNumberShouldMatch
	SHOULD = Occur("")

	// MUST_NOT
	// Use this operator for clauses that must not appear in the matching documents.
	// Note that it is not possible to search for queries that only consist of a MUST_NOT clause.
	// These clauses do not contribute to the score of documents.
	MUST_NOT = Occur("-")
)
