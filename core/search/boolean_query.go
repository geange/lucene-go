package search

import (
	"errors"

	"github.com/geange/lucene-go/core/util/structure"
)

var (
	maxClauseCount = 1024
)

// BooleanQuery
// A Query that matches documents matching boolean combinations of other queries,
// e.g. TermQuerys, PhraseQuerys or other BooleanQuerys.
type BooleanQuery struct {
	minimumNumberShouldMatch int
	clauses                  []*BooleanClause
	clauseSets               map[Occur][]Query
}

func newBooleanQuery(minimumNumberShouldMatch int, clauses []*BooleanClause) *BooleanQuery {
	query := &BooleanQuery{
		minimumNumberShouldMatch: minimumNumberShouldMatch,
		clauses:                  clauses,
		clauseSets:               map[Occur][]Query{SHOULD: {}, MUST: {}, FILTER: {}, MUST_NOT: {}},
	}
	for _, clause := range clauses {
		key := clause.GetOccur()
		query.clauseSets[key] = append(query.clauseSets[key], clause.GetQuery())
	}
	return query
}

// GetMinimumNumberShouldMatch
// Gets the minimum number of the optional BooleanClauses which must be satisfied.
func (b *BooleanQuery) GetMinimumNumberShouldMatch() int {
	return b.minimumNumberShouldMatch
}

// Clauses
// Return a list of the clauses of this BooleanQuery.
func (b *BooleanQuery) Clauses() []*BooleanClause {
	return b.clauses
}

// GetClauses
// Return the collection of queries for the given BooleanClause.Occur.
func (b *BooleanQuery) GetClauses(occur Occur) []Query {
	return b.clauseSets[occur]
}

// Whether this query is a pure disjunction,
// ie. it only has SHOULD clauses and it is enough for a single clause to match for this boolean query to match.
func (b *BooleanQuery) isPureDisjunction() bool {
	return len(b.clauses) == len(b.GetClauses(SHOULD)) &&
		b.minimumNumberShouldMatch <= 1
}

func (b *BooleanQuery) Iterator() structure.Iterator[*BooleanClause] {
	return structure.NewArrayListArray(b.clauses).Iterator()
}

// GetMaxClauseCount
// Return the maximum number of clauses permitted, 1024 by default.
// Attempts to add more than the permitted number of clauses cause BooleanQuery.TooManyClauses
// to be thrown.
//
// See Also: setMaxClauseCount(int)
func GetMaxClauseCount() int {
	return maxClauseCount
}

// Set the maximum number of clauses permitted per BooleanQuery. Default value is 1024.
func setMaxClauseCount(v int) {
	maxClauseCount = v
}

// BooleanQueryBuilder A builder for boolean queries.
type BooleanQueryBuilder struct {
	minimumNumberShouldMatch int
	clauses                  []*BooleanClause
	errs                     []error
}

// SetMinimumNumberShouldMatch
// Specifies a minimum number of the optional BooleanClauses which must be satisfied.
// By default no optional clauses are necessary for a match (unless there are no required clauses).
// If this method is used, then the specified number of clauses is required.
// Use of this method is totally independent of specifying that any specific clauses are required (or prohibited).
// This number will only be compared against the number of matching optional clauses.
// Params: min – the number of optional clauses that must match
func (b *BooleanQueryBuilder) SetMinimumNumberShouldMatch(min int) *BooleanQueryBuilder {
	b.minimumNumberShouldMatch = min
	return b
}

// Add
// a new clause to this BooleanQuery.Builder.
// Note that the order in which clauses are added does not have any impact on matching documents or query performance.
// Throws: BooleanQuery.TooManyClauses – if the new number of clauses exceeds the maximum clause number
func (b *BooleanQueryBuilder) Add(clause *BooleanClause) *BooleanQueryBuilder {
	if len(b.clauses) >= maxClauseCount {
		b.errs = append(b.errs, errors.New("TooManyClauses"))
		return b
	}
	b.clauses = append(b.clauses, clause)
	return b
}

// AddQuery
// a new clause to this BooleanQuery.Builder.
// Note that the order in which clauses are added does not have any impact on matching documents or query performance.
// Throws: BooleanQuery.TooManyClauses – if the new number of clauses exceeds the maximum clause number
func (b *BooleanQueryBuilder) AddQuery(query Query, occur Occur) *BooleanQueryBuilder {
	return b.Add(NewBooleanClause(query, occur))
}

func (b *BooleanQueryBuilder) Build() *BooleanQuery {
	return newBooleanQuery(b.minimumNumberShouldMatch, b.clauses)
}
