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

func NewAutomatonQuery(term *index.Term, auto *automaton.Automaton, determinizeWorkLimit int, isBinary bool) *AutomatonQuery {
	return &AutomatonQuery{
		field:             term.Field(),
		automaton:         auto,
		term:              term,
		automatonIsBinary: isBinary,
		compiled:          automaton.NewCompiledAutomaton(auto, nil, true, determinizeWorkLimit, isBinary),
	}
}

func (r *AutomatonQuery) GetField() string {
	return r.field
}

func (r *AutomatonQuery) GetTermsEnum(terms index.Terms, atts *tokenattributes.AttributeSource) (index.TermsEnum, error) {
	return r.compiled.GetTermsEnum(terms)
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
	return r, nil
}

func (r *AutomatonQuery) Visit(visitor QueryVisitor) (err error) {
	if visitor.AcceptField(r.field) {
		visit(r.compiled, visitor, r, r.field)
	}
	return nil
}

func visit(auto *automaton.CompiledAutomaton, visitor QueryVisitor, parent Query, field string) {
	if visitor.AcceptField(field) {
		switch auto.Type() {
		case automaton.AUTOMATON_TYPE_NORMAL:
			visitor.ConsumeTermsMatching(parent, field, auto.RunAutomaton)
		case automaton.AUTOMATON_TYPE_NONE:
		case automaton.AUTOMATON_TYPE_ALL:
			visitor.ConsumeTermsMatching(parent, field, func() *automaton.ByteRunAutomaton {
				return automaton.NewByteRunAutomaton(automaton.MakeAnyString())
			})
		case automaton.AUTOMATON_TYPE_SINGLE:
			visitor.ConsumeTerms(parent, index.NewTerm(field, auto.Term()))
		}
	}
}
