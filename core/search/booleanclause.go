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
	return OccurMustNot == b.occur
}

func (b *BooleanClause) IsRequired() bool {
	return b.occur == OccurMust || b.occur == OccurFilter
}

func (b *BooleanClause) IsScoring() bool {
	return b.occur == OccurMust || b.occur == OccurShould
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

func OccurValues() []Occur {
	return []Occur{
		OccurMust, OccurFilter, OccurShould, OccurMustNot,
	}
}

const (
	// OccurMust
	// Use this operator for clauses that must appear in the matching documents.
	// 等同于 AND
	OccurMust = Occur("+")

	// OccurFilter
	// Like OccurMust except that these clauses do not participate in scoring.
	OccurFilter = Occur("#")

	// OccurShould
	// Use this operator for clauses that should appear in the matching documents.
	// For a BooleanQuery with no OccurMust clauses one or more OccurShould clauses must match
	// a document for the BooleanQuery to match.
	// See Also: BooleanQuery.BooleanQueryBuilder.setMinimumNumberShouldMatch
	// 等同于 OR
	OccurShould = Occur("")

	// OccurMustNot
	// Use this operator for clauses that must not appear in the matching documents.
	// Note that it is not possible to search for queries that only consist of a OccurMustNot clause.
	// These clauses do not contribute to the score of documents.
	// 等同于 NOT
	OccurMustNot = Occur("-")
)
