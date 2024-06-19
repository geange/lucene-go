package search

import (
	"fmt"
	"github.com/geange/lucene-go/core/interface/search"
)

// BooleanClause
// A clause in a BooleanQuery.
type BooleanClause struct {
	// The query whose matching documents are combined by the boolean query.
	query search.Query

	occur search.Occur
}

func NewBooleanClause(query search.Query, occur search.Occur) *BooleanClause {
	return &BooleanClause{occur: occur, query: query}
}

func (b *BooleanClause) GetOccur() search.Occur {
	return b.occur
}

func (b *BooleanClause) GetQuery() search.Query {
	return b.query
}

func (b *BooleanClause) IsProhibited() bool {
	return search.OccurMustNot == b.occur
}

func (b *BooleanClause) IsRequired() bool {
	return b.occur == search.OccurMust || b.occur == search.OccurFilter
}

func (b *BooleanClause) IsScoring() bool {
	return b.occur == search.OccurMust || b.occur == search.OccurShould
}

func (b *BooleanClause) String() string {
	return fmt.Sprintf("%s %s", b.occur, b.query)
}
