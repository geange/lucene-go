package search

import "github.com/geange/lucene-go/core/index"

// TermInSetQuery
// Specialization for a disjunction over many terms that behaves like a ConstantScoreQuery over
// a BooleanQuery containing only BooleanClause.Occur.SHOULD clauses.
//
// For instance in the following example, both q1 and q2 would yield the same scores:
// Query q1 = new TermInSetQuery("field", new BytesRef("foo"), new BytesRef("bar"));
//
// BooleanQuery bq = new BooleanQuery();
// bq.add(new TermQuery(new Term("field", "foo")), Occur.SHOULD);
// bq.add(new TermQuery(new Term("field", "bar")), Occur.SHOULD);
// Query q2 = new ConstantScoreQuery(bq);
//
// When there are few terms, this query executes like a regular disjunction. However,
// when there are many terms, instead of merging iterators on the fly, it will populate a
// bit set with matching docs and return a Scorer over this bit set.
//
// NOTE: This query produces scores that are equal to its boost
type TermInSetQuery struct {
	field    string
	termData *index.PrefixCodedTerms
}
