package automaton

import (
	"errors"
	"math"
	"sync/atomic"
	"unicode/utf8"
)

// CompiledAutomaton Immutable class holding compiled details for a given Automaton. The Automaton is
// deterministic, must not have dead states but is not necessarily minimal.
type CompiledAutomaton struct {
	// If simplify is true this will be the "simplified" types; else, this is NORMAL
	_type int

	// For CompiledAutomaton.AUTOMATON_TYPE.SINGLE this is the singleton term.
	term []byte

	// Matcher for quickly determining if a byte[] is accepted. only valid for CompiledAutomaton.AUTOMATON_TYPE.NORMAL.
	runAutomaton *ByteRunAutomaton

	// Two dimensional array of transitions, indexed by state number for traversal. The state numbering is
	// consistent with runAutomaton. Only valid for CompiledAutomaton.AUTOMATON_TYPE.NORMAL.
	automaton *Automaton

	// Shared common suffix accepted by the automaton. Only valid for CompiledAutomaton.AUTOMATON_TYPE.NORMAL,
	// and only when the automaton accepts an infinite language. This will be null if the common prefix is length 0.
	commonSuffixRef []byte

	// Indicates if the automaton accepts a finite set of strings. Null if this was not computed. Only valid
	// for CompiledAutomaton.AUTOMATON_TYPE.NORMAL.
	finite *atomic.Bool

	// Which state, if any, accepts all suffixes, else -1.
	sinkState int

	transition *Transition
}

// NewCompiledAutomaton
// Create this. If finite is null, we use Operations.isFinite to determine whether it is finite. If simplify is true, we run possibly expensive operations to determine if the automaton is one the cases in CompiledAutomaton.AUTOMATON_TYPE. If simplify requires determinizing the automaton then at most determinizeWorkLimit effort will be spent. Any more than that will cause a TooComplexToDeterminizeException.
func NewCompiledAutomaton(automaton *Automaton, finite *atomic.Bool, simplify bool,
	determinizeWorkLimit int, isBinary bool) *CompiledAutomaton {

	this := &CompiledAutomaton{}

	if automaton.GetNumStates() == 0 {
		automaton = NewAutomaton()
		automaton.CreateState()
	}

	if simplify {

		// Test whether the automaton is a "simple" form and
		// if so, don't create a runAutomaton.  Note that on a
		// large automaton these tests could be costly:

		if IsEmptyAutomaton(automaton) {
			// matches nothing
			this._type = AUTOMATON_TYPE_NONE
			this.term = nil
			this.commonSuffixRef = nil
			this.runAutomaton = nil
			this.automaton = nil
			this.finite = nil
			this.sinkState = -1
			return this
		}

		var isTotal bool

		// NOTE: only approximate, because automaton may not be minimal:
		if isBinary {
			isTotal = IsTotalAutomatonRange(automaton, 0, 0xff)
		} else {
			isTotal = IsTotalAutomaton(automaton)
		}

		if isTotal {
			// matches all possible strings
			this._type = AUTOMATON_TYPE_ALL
			this.term = nil
			this.commonSuffixRef = nil
			this.runAutomaton = nil
			this.automaton = nil
			this.finite = nil
			this.sinkState = -1
			return this
		}

		automaton = DeterminizeAutomaton(automaton, determinizeWorkLimit)

		singleton, _ := GetSingletonAutomaton(automaton)

		if singleton != nil {
			// matches a fixed string
			this._type = AUTOMATON_TYPE_SINGLE
			this.commonSuffixRef = nil
			this.runAutomaton = nil
			this.automaton = nil
			this.finite = nil

			if isBinary {
				this.term, _ = intsToBytes(singleton)
			} else {
				this.term, _ = unicodeIntsToBytes(singleton)
			}
			this.sinkState = -1
			return this
		}
	}

	this._type = AUTOMATON_TYPE_NORMAL
	this.term = nil

	if finite == nil {
		this.finite = IsFiniteAutomaton(automaton)
	} else {
		this.finite = finite
	}

	var binary *Automaton
	if isBinary {
		// Caller already built binary automaton themselves, e.g. PrefixQuery
		// does this since it can be provided with a binary (not necessarily
		// UTF8!) term:
		binary = automaton
	} else {
		// Incoming automaton is unicode, and we must convert to UTF8 to match what's in the index:
		// FIXME
	}

	// compute a common suffix for infinite DFAs, this is an optimization for "leading wildcard"
	// so don't burn cycles on it if the DFA is finite, or largeish
	if this.finite.Load() || automaton.GetNumStates()+automaton.GetNumTransitions() > 1000 {
		this.commonSuffixRef = nil
	} else {
		suffix := GetCommonSuffixBytesRef(binary)
		if len(suffix) == 0 {
			this.commonSuffixRef = nil
		} else {
			this.commonSuffixRef = suffix
		}
	}

	// This will determinize the binary automaton for us:
	this.runAutomaton = NewByteRunAutomatonV1(binary, true, determinizeWorkLimit)
	this.automaton = this.runAutomaton.automaton

	// TODO: this is a bit fragile because if the automaton is not minimized there could be more than 1 sink state but this-prefix will fail
	// to run for those:
	this.sinkState = findSinkState(this.automaton)
	return this
}

func findSinkState(automaton *Automaton) int {
	numStates := automaton.GetNumStates()
	t := NewTransition()
	foundState := -1
	for s := 0; s < numStates; s++ {
		if automaton.IsAccept(s) {
			count := automaton.InitTransition(s, t)
			isSinkState := false
			for i := 0; i < count; i++ {
				automaton.GetNextTransition(t)
				if t.Dest == s && t.Min == 0 && t.Max == 0xff {
					isSinkState = true
					break
				}
			}
			if isSinkState {
				foundState = s
				break
			}
		}
	}
	return foundState
}

func intsToBytes(values []int) ([]byte, error) {
	bs := make([]byte, len(values))
	for i, value := range values {
		if value < 0 || value > 255 {
			return nil, errors.New("out-of-bounds for byte")
		}
		bs[i] = byte(value)
	}
	return bs, nil
}

func unicodeIntsToBytes(values []int) ([]byte, error) {
	bs := make([]byte, len(values)*3)
	i := 0
	for _, value := range values {
		if value < 0 || value > math.MaxInt32 {
			return nil, errors.New("out-of-bounds for byte")
		}
		size := utf8.EncodeRune(bs[i:], rune(value))
		i += size
	}
	return bs[:i], nil
}

func (r *CompiledAutomaton) Type() int {
	return r._type
}

func (r *CompiledAutomaton) Term() []byte {
	return r.term
}

func (r *CompiledAutomaton) RunAutomaton() *ByteRunAutomaton {
	return r.runAutomaton
}

//func (r *CompiledAutomaton) GetTermsEnum(terms index.Terms) (index.TermsEnum, error) {
//	switch r._type {
//	case AUTOMATON_TYPE_NONE:
//		return index.EmptyTermsEnum, nil
//	case AUTOMATON_TYPE_ALL:
//		return terms.Iterator()
//	case AUTOMATON_TYPE_SINGLE:
//		it, err := terms.Iterator()
//		if err != nil {
//			return nil, err
//		}
//		return index.NewSingleTermsEnum(it, r.term), nil
//	case AUTOMATON_TYPE_NORMAL:
//		return terms.Intersect(r, nil)
//	default:
//		return nil, errors.New("unhandled case")
//	}
//}

const (
	AUTOMATON_TYPE_NONE   = iota // Automaton that accepts no strings.
	AUTOMATON_TYPE_ALL           // Automaton that accepts all possible strings.
	AUTOMATON_TYPE_SINGLE        // Automaton that accepts only a single fixed string.
	AUTOMATON_TYPE_NORMAL        // Catch-all for any other automata.
)
