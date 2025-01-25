package search

import (
	"fmt"

	"github.com/geange/lucene-go/core/interface/index"
)

// BooleanClause
// A clause in a BooleanQuery.
type BooleanClause struct {
	// The query whose matching documents are combined by the boolean query.
	query index.Query

	occur index.Occur
}

func NewBooleanClause(query index.Query, occur index.Occur) *BooleanClause {
	return &BooleanClause{occur: occur, query: query}
}

func (b *BooleanClause) GetOccur() index.Occur {
	return b.occur
}

func (b *BooleanClause) GetQuery() index.Query {
	return b.query
}

func (b *BooleanClause) IsProhibited() bool {
	return index.OccurMustNot == b.occur
}

func (b *BooleanClause) IsRequired() bool {
	return b.occur == index.OccurMust || b.occur == index.OccurFilter
}

func (b *BooleanClause) IsScoring() bool {
	return b.occur == index.OccurMust || b.occur == index.OccurShould
}

func (b *BooleanClause) String() string {
	return fmt.Sprintf("%s %s", b.occur, b.query)
}
