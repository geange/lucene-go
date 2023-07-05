package search

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/tokenattributes"
	"github.com/geange/lucene-go/core/util/automaton"
)

var _ MultiTermQuery = &AutomatonQuery{}

// AutomatonQuery
// A Query that will match terms against a finite-state machine.
//
// This query will match documents that contain terms accepted by a given finite-state machine.
// The automaton can be constructed with the org.apache.lucene.util.automaton API. Alternatively,
// it can be created from a regular expression with RegexpQuery or from the standard Lucene wildcard
// syntax with WildcardQuery.
//
// When the query is executed, it will create an equivalent DFA of the finite-state machine,
// and will enumerate the term dictionary in an intelligent way to reduce the number of comparisons.
// For example: the regular expression of [dl]og? will make approximately four comparisons: do, dog, lo, and log.
// lucene.experimental
type AutomatonQuery struct {
	field string

	//  the automaton to match index terms against
	automaton *automaton.Automaton
	compiled  *automaton.CompiledAutomaton

	// term containing the field, and possibly some pattern structure
	term *index.Term

	automatonIsBinary bool

	rewriteMethod RewriteMethod
}

func (r *AutomatonQuery) GetField() string {
	return r.field
}

func (r *AutomatonQuery) GetTermsEnum(terms index.Terms, atts *tokenattributes.AttributeSource) (index.TermsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (r *AutomatonQuery) GetRewriteMethod() RewriteMethod {
	//TODO implement me
	panic("implement me")
}

func (r *AutomatonQuery) SetRewriteMethod(method RewriteMethod) {
	//TODO implement me
	panic("implement me")
}

func (r *AutomatonQuery) String(field string) string {
	//TODO implement me
	panic("implement me")
}

func (r *AutomatonQuery) CreateWeight(searcher *IndexSearcher, scoreMode *ScoreMode, boost float64) (Weight, error) {
	//TODO implement me
	panic("implement me")
}

func (r *AutomatonQuery) Rewrite(reader index.IndexReader) (Query, error) {
	//TODO implement me
	panic("implement me")
}

func (r *AutomatonQuery) Visit(visitor QueryVisitor) (err error) {
	//TODO implement me
	panic("implement me")
}
