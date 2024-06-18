package search

import (
	"errors"
	index3 "github.com/geange/lucene-go/core/interface/index"
	index2 "github.com/geange/lucene-go/core/types"

	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util/attribute"
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
	term *index2.Term

	automatonIsBinary bool

	rewriteMethod RewriteMethod
}

func NewAutomatonQuery(term *index2.Term, auto *automaton.Automaton, determinizeWorkLimit int, isBinary bool) *AutomatonQuery {
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

func (r *AutomatonQuery) GetTermsEnum(terms index3.Terms, atts *attribute.Source) (index3.TermsEnum, error) {
	return GetTermsEnum(r.compiled, terms)
}

func GetTermsEnum(r *automaton.CompiledAutomaton, terms index3.Terms) (index3.TermsEnum, error) {
	switch r.Type() {
	case automaton.AUTOMATON_TYPE_NONE:
		return index.EmptyTermsEnum, nil
	case automaton.AUTOMATON_TYPE_ALL:
		return terms.Iterator()
	case automaton.AUTOMATON_TYPE_SINGLE:
		it, err := terms.Iterator()
		if err != nil {
			return nil, err
		}
		return index.NewSingleTermsEnum(it, r.Term()), nil
	case automaton.AUTOMATON_TYPE_NORMAL:
		return terms.Intersect(r, nil)
	default:
		return nil, errors.New("unhandled case")
	}
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

func (r *AutomatonQuery) CreateWeight(searcher *IndexSearcher, scoreMode ScoreMode, boost float64) (Weight, error) {
	return nil, errors.New("implement me")
}

func (r *AutomatonQuery) Rewrite(reader index3.IndexReader) (Query, error) {
	return r, nil
}

func (r *AutomatonQuery) Visit(visitor QueryVisitor) error {
	if visitor.AcceptField(r.field) {
		if err := visit(r.compiled, visitor, r, r.field); err != nil {
			return err
		}
	}
	return nil
}

func visit(auto *automaton.CompiledAutomaton, visitor QueryVisitor, parent Query, field string) error {
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
			visitor.ConsumeTerms(parent, index2.NewTerm(field, auto.Term()))
		default:
			return errors.New("unhandled case")
		}
	}
	return nil
}
