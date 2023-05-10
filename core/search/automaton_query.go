package search

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util/automaton"
)

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
}
