package index

import (
	"bytes"
	"github.com/geange/lucene-go/core/util/automaton"
)

// AutomatonTermsEnum A FilteredTermsEnum that enumerates terms based upon what is accepted by a DFA.
// The algorithm is such:
// As long as matches are successful, keep reading sequentially.
// When a match fails, skip to the next string in lexicographic order that does not enter a reject state.
// The algorithm does not attempt to actually skip to the next string that is completely accepted.
// This is not possible when the language accepted by the FSM is not finite (i.e. * operator).
// lucene.internal
type AutomatonTermsEnum struct {
	*FilteredTermsEnumDefault

	// a tableized array-based form of the DFA
	runAutomaton *automaton.ByteRunAutomaton

	// common suffix of the automaton
	commonSuffixRef []byte

	// true if the automaton accepts a finite language
	finite bool

	// array of sorted transitions for each state, indexed by state number
	automaton *automaton.Automaton

	// Used for visited state tracking: each short records gen when we last
	// visited the state; we use gens to avoid having to clear
	visited []int

	curGen int

	// the reference used for seeking forwards through the term dictionary
	seekBytesRef any

	// true if we are enumerating an infinite portion of the DFA.
	// in this case it is faster to drive the query based on the terms dictionary.
	// when this is true, linearUpperBound indicate the end of range
	// of terms where we should simply do sequential reads instead.
	linear bool

	linearUpperBound []byte

	transition *automaton.Transition

	savedStates any
}

// Records the given state has been visited.
func (a *AutomatonTermsEnum) setVisited(state int) {
	if !a.finite {
		a.visited[state] = state
	}
}

// Indicates whether the given state has been visited.
func (a *AutomatonTermsEnum) isVisited(state int) bool {
	return !a.finite && a.visited[state] == a.curGen
}

func (a *AutomatonTermsEnum) Accept(term []byte) (AcceptStatus, error) {
	if len(a.commonSuffixRef) == 0 || bytes.HasPrefix(term, a.commonSuffixRef) {
		if a.runAutomaton.Run(term) {
			if a.linear {
				return ACCEPT_STATUS_YES, nil
			}
			return ACCEPT_STATUS_YES_AND_SEEK, nil
		}

		if a.linear && bytes.Compare(term, a.linearUpperBound) < 0 {
			return ACCEPT_STATUS_NO, nil
		}
		return ACCEPT_STATUS_NO_AND_SEEK, nil
	}

	if a.linear && bytes.Compare(term, a.linearUpperBound) < 0 {
		return ACCEPT_STATUS_NO, nil
	}
	return ACCEPT_STATUS_NO_AND_SEEK, nil
}
