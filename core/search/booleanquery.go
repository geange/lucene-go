package search

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/geange/gods-generic/lists/arraylist"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/interface/search"
)

var (
	maxClauseCount = 1024
)

var _ search.Query = &BooleanQuery{}

// BooleanQuery
// A Query that matches documents matching boolean combinations of other queries,
// e.g. TermQuerys, PhraseQuerys or other BooleanQuerys.
type BooleanQuery struct {
	minimumNumberShouldMatch int
	clauses                  []*BooleanClause
	clauseSets               map[search.Occur][]search.Query
}

func (b *BooleanQuery) String(field string) string {
	buf := new(bytes.Buffer)
	needParens := b.GetMinimumNumberShouldMatch() > 0
	if needParens {
		buf.WriteString("(")
	}

	for i, c := range b.clauses {
		buf.WriteString(c.GetOccur().String())

		subQuery := c.GetQuery()
		if _, ok := subQuery.(*BooleanQuery); ok {
			buf.WriteString("(")
			buf.WriteString(subQuery.String(field))
			buf.WriteString(")")
		} else {
			buf.WriteString(subQuery.String(field))
		}

		if i != len(b.clauses)-1 {
			buf.WriteString(" ")
		}
	}

	if needParens {
		buf.WriteString(")")
	}

	if b.GetMinimumNumberShouldMatch() > 0 {
		buf.WriteString("~")
		buf.WriteString(fmt.Sprintf("%d", b.GetMinimumNumberShouldMatch()))
	}
	return buf.String()
}

func (b *BooleanQuery) CreateWeight(searcher search.IndexSearcher, scoreMode search.ScoreMode, boost float64) (search.Weight, error) {
	query := b
	if scoreMode.NeedsScores() == false {
		booleanQuery, err := b.rewriteNoScoring()
		if err != nil {
			return nil, err
		}
		query = booleanQuery
	}
	return NewBooleanWeight(query, searcher, scoreMode, boost)
}

func (b *BooleanQuery) Rewrite(reader index.IndexReader) (search.Query, error) {
	if b.clauses == nil || len(b.clauses) == 0 {
		return nil, errors.New("empty BooleanQuery")
	}

	// optimize 1-clause queries
	if len(b.clauses) == 1 {
		c := b.clauses[0]
		query := c.GetQuery()
		if b.minimumNumberShouldMatch == 1 && c.GetOccur() == search.OccurShould {
			return query, nil
		}

		if b.minimumNumberShouldMatch == 0 {
			switch c.GetOccur() {
			case search.OccurShould:
			case search.OccurMust:
			case search.OccurFilter:
				return NewBoostQuery(NewConstantScoreQuery(query), 0)
			case search.OccurMustNot:
				return NewMatchNoDocsQuery("pure negative BooleanQuery"), nil
			default:
				return nil, errors.New("AssertionError")
			}
		}
	}

	// recursively rewrite
	{
		builder := NewBooleanQueryBuilder()
		builder.SetMinimumNumberShouldMatch(b.getMinimumNumberShouldMatch())
		actuallyRewritten := false
		for _, clause := range b.Clauses() {
			query := clause.GetQuery()
			rewritten, err := query.Rewrite(reader)
			if err != nil {
				return nil, err
			}
			if rewritten != query {
				// rewrite clause
				actuallyRewritten = true
				builder.AddQuery(rewritten, clause.GetOccur())
			} else {
				// leave as-is
				builder.Add(clause)
			}
		}
		if actuallyRewritten {
			return builder.Build()
		}
	}

	// remove duplicate OccurFilter and OccurMustNot clauses
	clauseCount := 0
	for _, queries := range b.clauseSets {
		clauseCount += len(queries)
	}

	if clauseCount != len(b.clauses) {
		// since clauseSets implicitly deduplicates OccurFilter and OccurMustNot
		// clauses, this means there were duplicates
		rewritten := NewBooleanQueryBuilder()
		rewritten.SetMinimumNumberShouldMatch(b.minimumNumberShouldMatch)

		for occur, queries := range b.clauseSets {
			for _, query := range queries {
				rewritten.AddQuery(query, occur)
			}
		}
		return rewritten.Build()
	}

	// Check whether some clauses are both required and excluded
	mustNotClauses := b.clauseSets[search.OccurMustNot]
	if len(mustNotClauses) > 0 {
		filter := make(map[search.Query]struct{})
		for _, v := range b.clauseSets[search.OccurFilter] {
			filter[v] = struct{}{}
		}

		for _, query := range mustNotClauses {
			if _, ok := filter[query]; ok {
				return NewMatchNoDocsQuery("OccurFilter or OccurMust clause also in OccurMustNot"), nil
			}

			if _, ok := query.(*MatchAllDocsQuery); ok {
				return NewMatchNoDocsQuery("OccurMustNot clause is MatchAllDocsQuery"), nil
			}
		}
	}

	// remove OccurFilter clauses that are also OccurMust clauses or that match all documents
	if len(b.clauseSets[search.OccurFilter]) > 0 {
		filters := make(map[search.Query]struct{})
		for _, v := range b.clauseSets[search.OccurFilter] {
			filters[v] = struct{}{}
		}

		modified := false
		if len(filters) > 1 || len(b.clauseSets[search.OccurMust]) > 0 {
			keys := make([]search.Query, 0)
			for query := range filters {
				if _, ok := query.(*MatchAllDocsQuery); ok {
					keys = append(keys, query)
				}
			}

			if len(keys) > 0 {
				modified = true

				for _, key := range keys {
					delete(filters, key)
				}
			}
		}

		for _, query := range b.clauseSets[search.OccurMust] {
			if _, ok := filters[query]; ok {
				modified = true

				delete(filters, query)
			}
		}

		if modified {
			builder := NewBooleanQueryBuilder()
			builder.SetMinimumNumberShouldMatch(b.getMinimumNumberShouldMatch())
			for _, clause := range b.clauses {
				if clause.GetOccur() != search.OccurFilter {
					builder.Add(clause)
				}
			}

			for query := range filters {
				builder.AddQuery(query, search.OccurFilter)
			}

			return builder.Build()
		}
	}

	// convert OccurFilter clauses that are also OccurShould clauses to OccurMust clauses
	if len(b.clauseSets[search.OccurShould]) > 0 && len(b.clauseSets[search.OccurFilter]) > 0 {
		filters := b.clauseSets[search.OccurFilter]
		shoulds := b.clauseSets[search.OccurShould]

		shouldsMap := make(map[search.Query]struct{})
		for _, query := range shoulds {
			shouldsMap[query] = struct{}{}
		}

		intersection := make(map[search.Query]struct{})
		for _, query := range filters {
			if _, ok := shouldsMap[query]; !ok {
				continue
			}
			intersection[query] = struct{}{}
		}

		if len(intersection) > 0 {
			builder := NewBooleanQueryBuilder()
			minShouldMatch := b.getMinimumNumberShouldMatch()

			for _, clause := range b.clauses {
				if _, ok := intersection[clause.GetQuery()]; ok {
					if clause.GetOccur() == search.OccurShould {
						builder.Add(NewBooleanClause(clause.GetQuery(), search.OccurMust))
						minShouldMatch--
					}
				} else {
					builder.Add(clause)
				}
			}

			builder.SetMinimumNumberShouldMatch(max(0, minShouldMatch))
			return builder.Build()
		}
	}

	// Deduplicate OccurShould clauses by summing up their boosts
	if len(b.clauseSets[search.OccurShould]) > 0 && b.minimumNumberShouldMatch <= 1 {
		shouldClauses := make(map[search.Query]float64)

		for _, query := range b.clauseSets[search.OccurShould] {
			boost := 1.0
			for {
				bq, ok := query.(*BoostQuery)
				if !ok {
					break
				}
				boost *= bq.GetBoost()
				query = bq.GetQuery()
			}
			shouldClauses[query] += boost
		}

		if len(shouldClauses) != len(b.clauseSets[search.OccurShould]) {
			builder := NewBooleanQueryBuilder()
			builder.SetMinimumNumberShouldMatch(b.minimumNumberShouldMatch)

			for query, boost := range shouldClauses {
				if boost != 1 {
					var err error
					query, err = NewBoostQuery(query, boost)
					if err != nil {
						return nil, err
					}
				}
				builder.AddQuery(query, search.OccurShould)
			}

			for _, clause := range b.clauses {
				if clause.GetOccur() != search.OccurShould {
					builder.Add(clause)
				}
			}
			return builder.Build()
		}
	}

	// Deduplicate OccurMust clauses by summing up their boosts
	if len(b.clauseSets[search.OccurMust]) > 0 {
		mustClauses := make(map[search.Query]float64)

		for _, query := range b.clauseSets[search.OccurMust] {
			boost := 1.0
			for {
				bq, ok := query.(*BoostQuery)
				if !ok {
					break
				}
				boost *= bq.GetBoost()
				query = bq.GetQuery()
			}

			mustClauses[query] += boost
		}

		if len(mustClauses) != len(b.clauseSets[search.OccurMust]) {
			builder := NewBooleanQueryBuilder()
			builder.SetMinimumNumberShouldMatch(b.minimumNumberShouldMatch)
			for query, boost := range mustClauses {
				if boost != 1 {
					var err error
					query, err = NewBoostQuery(query, boost)
					if err != nil {
						return nil, err
					}
				}
				builder.AddQuery(query, search.OccurMust)
			}

			for _, clause := range b.clauses {
				if clause.GetOccur() != search.OccurMust {
					builder.Add(clause)
				}
			}
			return builder.Build()
		}
	}

	// Rewrite queries whose single scoring clause is a OccurMust clause on a
	// MatchAllDocsQuery to a ConstantScoreQuery
	{
		musts := b.clauseSets[search.OccurMust]
		filters := b.clauseSets[search.OccurFilter]
		if len(musts) == 1 && len(filters) > 0 {
			must := musts[0]
			boost := 1.0

			if boostQuery, ok := must.(*BoostQuery); ok {
				must = boostQuery.GetQuery()
				boost = boostQuery.GetBoost()
			}

			if _, ok := must.(*MatchAllDocsQuery); ok {
				// our single scoring clause matches everything: rewrite to a CSQ on the filter
				// ignore OccurShould clause for now
				builder := NewBooleanQueryBuilder()
				for _, clause := range b.clauses {
					switch clause.GetOccur() {
					case search.OccurFilter:
					case search.OccurMustNot:
						builder.Add(clause)
					default:
					}
				}

				var rewritten search.Query
				var err error
				rewritten, err = builder.Build()
				if err != nil {
					return nil, err
				}
				rewritten = NewConstantScoreQuery(rewritten)
				if boost != 1 {
					rewritten, err = NewBoostQuery(rewritten, boost)
					if err != nil {
						return nil, err
					}
				}

				// now add back the OccurShould clauses
				builder = NewBooleanQueryBuilder()
				builder.SetMinimumNumberShouldMatch(b.GetMinimumNumberShouldMatch())
				builder.AddQuery(rewritten, search.OccurMust)

				for _, query := range b.clauseSets[search.OccurShould] {
					builder.AddQuery(query, search.OccurShould)
				}
				rewritten, err = builder.Build()
				if err != nil {
					return nil, err
				}
				return rewritten, nil
			}
		}
	}

	// Flatten nested disjunctions, this is important for block-max WAND to perform well
	if b.minimumNumberShouldMatch <= 1 {
		builder := NewBooleanQueryBuilder()
		builder.SetMinimumNumberShouldMatch(b.minimumNumberShouldMatch)
		actuallyRewritten := false

		for _, clause := range b.clauses {
			query := clause.GetQuery()
			if innerQuery, ok := query.(*BooleanQuery); clause.GetOccur() == search.OccurShould && ok {
				if innerQuery.isPureDisjunction() {
					actuallyRewritten = true
					for _, innerClause := range innerQuery.clauses {
						builder.Add(innerClause)
					}
				} else {
					builder.Add(clause)
				}
			} else {
				builder.Add(clause)
			}
		}
		if actuallyRewritten {
			return builder.Build()
		}
	}

	return b, nil
}

func (b *BooleanQuery) Visit(visitor search.QueryVisitor) error {
	sub := visitor.GetSubVisitor(search.OccurMust, b)
	for occur, queries := range b.clauseSets {
		if len(queries) > 0 {
			if occur == search.OccurMust {
				for _, q := range b.clauseSets[occur] {
					if err := q.Visit(sub); err != nil {
						return err
					}
				}
			} else {
				vt := sub.GetSubVisitor(occur, b)
				for _, q := range b.clauseSets[occur] {
					if err := q.Visit(vt); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func newBooleanQuery(minimumNumberShouldMatch int, clauses []*BooleanClause) *BooleanQuery {
	query := &BooleanQuery{
		minimumNumberShouldMatch: minimumNumberShouldMatch,
		clauses:                  clauses,
		clauseSets:               map[search.Occur][]search.Query{search.OccurShould: {}, search.OccurMust: {}, search.OccurFilter: {}, search.OccurMustNot: {}},
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
func (b *BooleanQuery) GetClauses(occur search.Occur) []search.Query {
	return b.clauseSets[occur]
}

// Whether this query is a pure disjunction,
// ie. it only has OccurShould clauses and it is enough for a single clause to match for this boolean query to match.
func (b *BooleanQuery) isPureDisjunction() bool {
	return len(b.clauses) == len(b.GetClauses(search.OccurShould)) &&
		b.minimumNumberShouldMatch <= 1
}

func (b *BooleanQuery) Iterator() arraylist.Iterator[*BooleanClause] {
	return arraylist.New[*BooleanClause](b.clauses...).Iterator()
}

func (b *BooleanQuery) rewriteNoScoring() (*BooleanQuery, error) {
	keepShould := b.GetMinimumNumberShouldMatch() > 0 ||
		(len(b.clauseSets[search.OccurMust])+len(b.clauseSets[search.OccurFilter]) == 0)

	if len(b.clauseSets[search.OccurMust]) == 0 && keepShould {
		return b, nil
	}

	newQuery := NewBooleanQueryBuilder()
	newQuery.SetMinimumNumberShouldMatch(b.GetMinimumNumberShouldMatch())

	for _, clause := range b.clauses {
		switch clause.GetOccur() {
		case search.OccurMust:
			newQuery.AddQuery(clause.GetQuery(), search.OccurFilter)
		case search.OccurShould:
			if keepShould {
				newQuery.Add(clause)
			}
		default:
			newQuery.Add(clause)
		}
	}
	return newQuery.Build()
}

func (b *BooleanQuery) getMinimumNumberShouldMatch() int {
	return b.minimumNumberShouldMatch
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

func NewBooleanQueryBuilder() *BooleanQueryBuilder {
	return &BooleanQueryBuilder{
		minimumNumberShouldMatch: 0,
		clauses:                  make([]*BooleanClause, 0),
		errs:                     make([]error, 0),
	}
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
func (b *BooleanQueryBuilder) AddQuery(query search.Query, occur search.Occur) *BooleanQueryBuilder {
	return b.Add(NewBooleanClause(query, occur))
}

func (b *BooleanQueryBuilder) Build() (*BooleanQuery, error) {
	if len(b.errs) != 0 {
		return nil, errors.Join(b.errs...)
	}
	return newBooleanQuery(b.minimumNumberShouldMatch, b.clauses), nil
}
