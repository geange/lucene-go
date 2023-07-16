package automaton

import (
	"errors"
	"sync/atomic"

	"github.com/bits-and-blooms/bitset"
	"github.com/geange/gods-generic/cmp"
	"github.com/geange/lucene-go/core/util/bitmixer"
	"github.com/geange/lucene-go/core/util/structure"
)

// DeterminizeAutomaton Determinizes the given automaton.
// Worst case complexity: exponential in number of states.
// Params: 	workLimit – Maximum amount of "work" that the powerset construction will spend before throwing
//
//	TooComplexToDeterminizeException. Higher numbers allow this operation to consume more memory and
//	CPU but allow more complex automatons. Use DEFAULT_DETERMINIZE_WORK_LIMIT as a decent default
//	if you don't otherwise know what to specify.
//
// Throws: TooComplexToDeterminizeException – if determinizing requires more than workLimit "effort"
func DeterminizeAutomaton(a *Automaton, workLimit int) *Automaton {
	if a.IsDeterministic() {
		return a
	}
	if a.GetNumStates() <= 1 {
		// Already determinized
		return a
	}

	// subset construction
	b := NewNewBuilder()

	// Same initial values and state will always have the same hashCode
	initialset := NewFrozenIntSet([]int{0}, bitmixer.Mix32(0)+1, 0)
	// Create state 0:
	b.CreateState()

	worklist := make([]*FrozenIntSet, 0)
	newstate := structure.NewMap[IntSet, int]()

	worklist = append(worklist, initialset)
	b.SetAccept(0, a.IsAccept(0))
	newstate.Put(initialset, 0)

	// TODO:

	return a
}

// IsEmptyAutomaton
// Returns true if the given automaton accepts no strings.
func IsEmptyAutomaton(a *Automaton) bool {
	if a.GetNumStates() == 0 {
		// Common case: no states
		return true
	}

	if a.IsAccept(0) == false && a.GetNumTransitionsWithState(0) == 0 {
		// Common case: just one initial state
		return true
	}
	if a.IsAccept(0) == true {
		// Apparently common case: it accepts the damned empty string
		return false
	}

	workList := make([]int, 0)
	seen := bitset.New(uint(a.GetNumStates()))
	workList = append(workList, 0)
	seen.Set(0)

	t := NewTransition()
	for len(workList) > 0 {
		state := workList[0]
		workList = workList[1:]

		if a.IsAccept(state) {
			return false
		}

		count := a.InitTransition(state, t)
		for i := 0; i < count; i++ {
			a.GetNextTransition(t)
			if seen.Test(uint(t.Dest)) == false {
				workList = append(workList, t.Dest)
				seen.Set(uint(t.Dest))
			}
		}
	}
	return true
}

// IsTotalAutomaton
// Returns true if the given automaton accepts all strings. The automaton must be minimized.
func IsTotalAutomaton(a *Automaton) bool {
	return IsTotalAutomatonRange(a, 0, 0x10FFFF)
}

// IsTotalAutomatonRange
// Returns true if the given automaton accepts all strings for the specified min/max range of the alphabet.
// The automaton must be minimized.
func IsTotalAutomatonRange(a *Automaton, minAlphabet, maxAlphabet int) bool {
	if a.IsAccept(0) && a.GetNumTransitionsWithState(0) == 1 {
		t := NewTransition()
		a.getTransition(0, 0, t)
		return t.Dest == 0 && t.Min == minAlphabet && t.Max == maxAlphabet
	}
	return false
}

func GetSingletonAutomaton(a *Automaton) ([]int, error) {
	if a.IsDeterministic() == false {
		return nil, errors.New("input automaton must be deterministic")
	}

	ints := make([]int, 0)
	visited := make(map[int]struct{})
	s := 0
	t := NewTransition()
	for {
		visited[s] = struct{}{}

		if a.IsAccept(s) == false {
			if a.GetNumTransitionsWithState(s) == 1 {
				a.getTransition(s, 0, t)
				if _, ok := visited[t.Dest]; t.Min == t.Max && ok {
					ints = append(ints, t.Min)
					s = t.Dest
					continue
				}
			}
		} else if a.GetNumTransitionsWithState(s) == 0 {
			return ints, nil
		}

		// Automaton accepts more than one string:
		return nil, nil
	}
}

func IsFiniteAutomaton(a *Automaton) *atomic.Bool {
	flag := &atomic.Bool{}

	if a.GetNumStates() == 0 {
		flag.Store(true)
		return flag
	}

	b1 := bitset.New(uint(a.GetNumStates()))
	b2 := bitset.New(uint(a.GetNumStates()))

	return isFinite(NewTransition(), a, 0, b1, b2, 0)
}

// Checks whether there is a loop containing state. (This is sufficient since there are never transitions to dead states.)
// TODO: not great that this is recursive... in theory a
// large automata could exceed java's stack so the maximum level of recursion is bounded to 1000
func isFinite(scratch *Transition, a *Automaton, state int, path, visited *bitset.BitSet, level int) *atomic.Bool {
	flag := &atomic.Bool{}

	// if (level > MAX_RECURSION_LEVEL) {
	//      throw new IllegalArgumentException("input automaton is too large: " +  level);
	//    }
	path.Set(uint(state))
	numTransitions := a.InitTransition(state, scratch)
	for t := 0; t < numTransitions; t++ {
		a.getTransition(state, t, scratch)
		if path.Test(uint(scratch.Dest)) || (!visited.Test(uint(scratch.Dest)) && !isFinite(scratch, a, scratch.Dest, path, visited, level+1).Load()) {
			flag.Store(false)
			return flag
		}
	}
	path.Clear(uint(state))
	visited.Set(uint(state))
	flag.Store(true)
	return flag
}

// GetCommonSuffixBytesRef
// Returns the longest BytesRef that is a suffix of all accepted strings. Worst case complexity: quadratic with the number of states+transitions.
// Returns: common suffix, which can be an empty (length 0) BytesRef (never null)
func GetCommonSuffixBytesRef(a *Automaton) []byte {
	// reverse the language of the automaton, then reverse its common prefix.
	panic("")
}

func reverse[T cmp.Ordered](ref []T) {
	i, j := 0, len(ref)-1
	for i < j {
		ref[i], ref[j] = ref[j], ref[i]
	}
}

func reverseAutomaton(a *Automaton) *Automaton {
	return reverseAutomatonIntSet(a, nil)
}

func RemoveDeadStates(a *Automaton) *Automaton {
	numStates := a.GetNumStates()
	liveSet := getLiveStates(a)

	mp := make([]int, numStates)

	result := NewAutomaton()
	for i := 0; i < numStates; i++ {
		if liveSet.Test(uint(i)) {
			mp[i] = result.CreateState()
			result.SetAccept(mp[i], a.IsAccept(i))
		}
	}

	t := NewTransition()

	for i := 0; i < numStates; i++ {
		if liveSet.Test(uint(i)) {
			numTransitions := a.InitTransition(i, t)
			// filter out transitions to dead states:
			for j := 0; j < numTransitions; j++ {
				a.GetNextTransition(t)
				if liveSet.Test(uint(t.Dest)) {
					result.AddTransition(mp[i], mp[t.Dest], t.Min, t.Max)
				}
			}
		}
	}

	result.finishState()
	//assert hasDeadStates(result) == false;
	return result
}

func getLiveStates(a *Automaton) *bitset.BitSet {
	live := getLiveStatesFromInitial(a)
	live.Union(getLiveStatesToAccept(a))
	return live
}

func getLiveStatesFromInitial(a *Automaton) *bitset.BitSet {
	panic("")
}

func getLiveStatesToAccept(a *Automaton) *bitset.BitSet {
	panic("")
}

func reverseAutomatonIntSet(a *Automaton, initialStates map[int]struct{}) *Automaton {
	if IsEmptyAutomaton(a) {
		return NewAutomaton()
	}

	numStates := a.GetNumStates()

	// Build a new automaton with all edges reversed
	builder := NewNewBuilder()

	// Initial node; we'll add epsilon transitions in the end:
	builder.CreateState()

	for s := 0; s < numStates; s++ {
		builder.CreateState()
	}

	// Old initial state becomes new accept state:
	builder.SetAccept(1, true)

	t := NewTransition()
	for s := 0; s < numStates; s++ {
		numTransitions := a.GetNumTransitionsWithState(s)
		a.InitTransition(s, t)
		for i := 0; i < numTransitions; i++ {
			a.GetNextTransition(t)
			builder.AddTransition(t.Dest+1, s+1, t.Min, t.Max)
		}
	}

	result := builder.Finish()

	s := 0
	acceptStates := a.getAcceptStates()
	for {
		if _, ok := acceptStates.NextSet(uint(s)); !(ok && s < numStates) {
			break
		}

		result.AddEpsilon(0, s+1)
		if initialStates != nil {
			initialStates[s+1] = struct{}{}
		}
		s++
	}

	result.finishState()

	return result
}
