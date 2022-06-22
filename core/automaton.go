package core

import (
	"fmt"
	"github.com/bits-and-blooms/bitset"
	"sort"
)

// Automaton Represents an automaton and all its states and transitions. States are integers and must be
// created using createState. Mark a state as an accept state using setAccept. Add transitions using
// addTransition. Each state must have all of its transitions added at once; if this is too restrictive
// then use Automaton.Builder instead. State 0 is always the initial state. Once a state is finished,
// either because you've starting adding transitions to another state or you call finishState, then that
// states transitions are sorted (first by min, then max, then dest) and reduced (transitions with adjacent
// labels going to the same dest are combined).
type Automaton struct {
	// Where we next write to the int[] states; this increments by 2 for each added state because we
	// pack a pointer to the transitions array and a count of how many transitions leave the state.
	nextState int

	// Where we next write to in int[] transitions; this increments by 3 for each added transition because
	// we pack min, max, dest in sequence.
	nextTransition int

	// Current state we are adding transitions to; the caller must add all transitions for this state
	// before moving onto another state.
	curState int

	// Index in the transitions array, where this states leaving transitions are stored, or -1 if this state has not added any transitions yet, followed by number of transitions.
	states []int

	isAccept *bitset.BitSet

	// Holds toState, min, max for each transition.
	transitions []int

	// True if no state has two transitions leaving with the same label.
	deterministic bool
}

// CreateState Create a new state.
func (r *Automaton) CreateState() int {
	r.growStates()
	state := r.nextState / 2
	r.states[r.nextState] = -1
	r.nextState += 2
	return state
}

// SetAccept Set or clear this state as an accept state.
func (r *Automaton) SetAccept(state int, accept bool) {
	r.isAccept.SetTo(uint(state), accept)
}

// Sugar to get all transitions for all states. This is object-heavy; it's better to iterate state by state instead.
func (r *Automaton) getSortedTransitions() [][]Transition {
	numStates := r.GetNumStates()
	transitions := make([][]Transition, numStates)

	for s := 0; s < numStates; s++ {
		numTransitions := r.GetNumTransitionsWithState(s)
		transitions[s] = make([]Transition, numTransitions)

		for t := 0; t < numTransitions; t++ {
			transition := Transition{}
			r.getTransition(s, t, &transition)
			transitions[s][t] = transition
		}
	}

	return transitions
}

// Returns accept states. If the bit is set then that state is an accept state.
func (r *Automaton) getAcceptStates() *bitset.BitSet {
	return r.isAccept
}

// IsAccept Returns true if this state is an accept state.
func (r *Automaton) IsAccept(state int) bool {
	return r.isAccept.Test(uint(state))
}

// AddTransition Add a new transition with the specified source, dest, min, max.
func (r *Automaton) AddTransition(source, dest, min, max int) error {
	//bounds := r.nextState / 2

	r.growTransitions()
	if r.curState != source {
		if r.curState != -1 {
			r.finishCurrentState()
		}

		// Move to next source:
		r.curState = source
		if r.states[2*r.curState] != -1 {
			return fmt.Errorf("from state (%d) already had transitions added", source)
		}
		r.states[2*r.curState] = r.nextTransition
	}

	r.transitions[r.nextTransition] = dest
	r.nextTransition++
	r.transitions[r.nextTransition] = min
	r.nextTransition++
	r.transitions[r.nextTransition] = max
	r.nextTransition++

	// Increment transition count for this state
	r.states[2*r.curState+1]++
	return nil
}

// AddEpsilon Add a [virtual] epsilon transition between source and dest. Dest state must already have all
// transitions added because this method simply copies those same transitions over to source.
func (r *Automaton) AddEpsilon(source, dest int) {
	t := Transition{}
	count := r.InitTransition(dest, &t)

	for i := 0; i < count; i++ {
		r.GetNextTransition(&t)
		_ = r.AddTransition(source, t.Dest, t.Min, t.Max)
	}

	if r.IsAccept(dest) {
		r.SetAccept(source, true)
	}
}

// Copy Copies over all states/transitions from other. The states numbers are sequentially assigned (appended).
func (r *Automaton) Copy(other *Automaton) {

	// Bulk copy and then fixup the state pointers:
	stateOffset := r.GetNumStates()
	r.states = Grow(r.states, r.nextState+other.nextState)
	copy(r.states[r.nextState:r.nextState+other.nextState], other.states)
	for i := 0; i < other.nextState; i += 2 {
		if r.states[r.nextState+i] != -1 {
			r.states[r.nextState+i] += r.nextTransition
		}
	}

	r.nextState += other.nextState
	otherNumStates := other.GetNumStates()
	otherAcceptStates := other.getAcceptStates()
	state := uint(0)

	for {
		if state < uint(otherNumStates) {
			if state, ok := otherAcceptStates.NextSet(state); ok {
				r.SetAccept(stateOffset+int(state), true)
				state++
				continue
			}
		}

		break
	}

	// Bulk copy and then fixup dest for each transition:
	r.transitions = Grow(r.transitions, r.nextTransition+other.nextTransition)
	copy(r.transitions[r.nextTransition:r.nextTransition+other.nextTransition], other.transitions)
	for i := 0; i < other.nextTransition; i += 3 {
		r.transitions[r.nextTransition+i] += stateOffset
	}
	r.nextTransition += other.nextTransition

	if other.deterministic == false {
		r.deterministic = false
	}
}

// Freezes the last state, sorting and reducing the transitions.
func (r *Automaton) finishCurrentState() {
	numTransitions := r.states[2*r.curState+1]

	offset := r.states[2*r.curState]
	start := offset / 3

	sort.Sort(&destMinMaxSorter{
		from:      start,
		to:        start + numTransitions,
		Automaton: r,
	})

	// Reduce any "adjacent" transitions:
	upto := 0
	min := -1
	max := -1
	dest := -1

	for i := 0; i < numTransitions; i++ {
		tDest := r.transitions[offset+3*i]
		tMin := r.transitions[offset+3*i+1]
		tMax := r.transitions[offset+3*i+2]

		if dest == tDest {
			if tMin <= max+1 {
				if tMax > max {
					max = tMax
				}
			} else {
				if dest != -1 {
					r.transitions[offset+3*upto] = dest
					r.transitions[offset+3*upto+1] = min
					r.transitions[offset+3*upto+2] = max
					upto++
				}
				min = tMin
				max = tMax
			}
		} else {
			if dest != -1 {
				r.transitions[offset+3*upto] = dest
				r.transitions[offset+3*upto+1] = min
				r.transitions[offset+3*upto+2] = max
				upto++
			}
			dest = tDest
			min = tMin
			max = tMax
		}
	}

	if dest != -1 {
		// Last transition
		r.transitions[offset+3*upto] = dest
		r.transitions[offset+3*upto+1] = min
		r.transitions[offset+3*upto+2] = max
		upto++
	}

	r.nextTransition -= (numTransitions - upto) * 3
	r.states[2*r.curState+1] = upto

	// Sort transitions by min/max/dest:
	sort.Sort(&minMaxDestSorter{
		from:      start,
		to:        start + upto,
		Automaton: r,
	})

	if r.deterministic && upto > 1 {
		lastMax := r.transitions[offset+2]
		for i := 1; i < upto; i++ {
			min = r.transitions[offset+3*i+1]
			if min <= lastMax {
				r.deterministic = false
				break
			}
			lastMax = r.transitions[offset+3*i+2]
		}
	}
}

// IsDeterministic Returns true if this automaton is deterministic (for ever state there is only one
// transition for each label).
func (r *Automaton) IsDeterministic() bool {
	return r.deterministic
}

// Finishes the current state; call this once you are done adding transitions for a state. This is automatically called if you start adding transitions to a new source state, but for the last state you add you need to this method yourself.
func (r *Automaton) finishState() {
	if r.curState != -1 {
		r.finishCurrentState()
		r.curState = -1
	}
}

// GetNumStates How many states this automaton has.
func (r *Automaton) GetNumStates() int {
	return r.nextState / 2
}

// GetNumTransitions How many transitions this automaton has.
func (r *Automaton) GetNumTransitions() int {
	return r.nextTransition / 3
}

// GetNumTransitionsWithState How many transitions this state has.
func (r *Automaton) GetNumTransitionsWithState(state int) int {
	count := r.states[2*state+1]
	if count == -1 {
		return 0
	}
	return count
}

func (r *Automaton) growStates() {
	if r.nextState+2 > len(r.states) {
		r.states = Grow(r.states, r.nextState+2)
	}
}

func (r *Automaton) growTransitions() {
	if r.nextTransition+3 > len(r.transitions) {
		r.transitions = Grow(r.transitions, r.nextTransition+3)
	}
}

// Sorts transitions by dest, ascending, then min label ascending, then max label ascending
type destMinMaxSorter struct {
	from, to int
	*Automaton
}

func (r *destMinMaxSorter) Len() int {
	return r.to - r.from
}

func (r *destMinMaxSorter) Less(i, j int) bool {
	iStart := 3 * i
	jStart := 3 * j

	iDest := r.transitions[iStart]
	jDest := r.transitions[jStart]

	// First dest:
	if iDest < jDest {
		return false
	} else if iDest > jDest {
		return true
	}

	// Then min:
	iMin := r.transitions[iStart+1]
	jMin := r.transitions[jStart+1]
	if iMin < jMin {
		return false
	} else if iMin > jMin {
		return true
	}

	// Then max:
	iMax := r.transitions[iStart+2]
	jMax := r.transitions[jStart+2]
	if iMax < jMax {
		return false
	} else if iMax > jMax {
		return true
	}

	return false
}

func (r *destMinMaxSorter) Swap(i, j int) {
	iStart, jStart := 3*i, 3*j
	r.swapOne(iStart, jStart)
	r.swapOne(iStart+1, jStart+1)
	r.swapOne(iStart+2, jStart+2)
}

func (r *destMinMaxSorter) swapOne(i, j int) {
	r.transitions[i], r.transitions[j] =
		r.transitions[j], r.transitions[i]
}

// Sorts transitions by min label, ascending, then max label ascending, then dest ascending
type minMaxDestSorter struct {
	from, to int
	*Automaton
}

func (r *minMaxDestSorter) Len() int {
	return r.to - r.from
}

func (r *minMaxDestSorter) Less(i, j int) bool {
	iStart := 3 * i
	jStart := 3 * j

	// First min:
	iMin := r.transitions[iStart+1]
	jMin := r.transitions[jStart+1]
	if iMin < jMin {
		return false
	} else if iMin > jMin {
		return true
	}

	// Then max:
	iMax := r.transitions[iStart+2]
	jMax := r.transitions[jStart+2]
	if iMax < jMax {
		return false
	} else if iMax > jMax {
		return true
	}

	// Then dest:
	iDest := r.transitions[iStart]
	jDest := r.transitions[jStart]
	if iDest < jDest {
		return false
	} else if iDest > jDest {
		return true
	}

	return false
}

func (r *minMaxDestSorter) Swap(i, j int) {
	iStart, jStart := 3*i, 3*j
	r.swapOne(iStart, jStart)
	r.swapOne(iStart+1, jStart+1)
	r.swapOne(iStart+2, jStart+2)
}

func (r *minMaxDestSorter) swapOne(i, j int) {
	r.transitions[i], r.transitions[j] =
		r.transitions[j], r.transitions[i]
}

// InitTransition Initialize the provided Transition to iterate through all transitions leaving the specified
// state. You must call getNextTransition to get each transition. Returns the number of transitions leaving
// this state.
func (r *Automaton) InitTransition(state int, t *Transition) int {
	t.Source = state
	t.TransitionUpto = r.states[2*state]
	return r.GetNumTransitionsWithState(state)
}

// GetNextTransition Iterate to the next transition after the provided one
func (r *Automaton) GetNextTransition(t *Transition) {
	t.Dest = r.transitions[t.TransitionUpto]
	t.TransitionUpto++
	t.Min = r.transitions[t.TransitionUpto]
	t.TransitionUpto++
	t.Max = r.transitions[t.TransitionUpto]
	t.TransitionUpto++
}

func (r *Automaton) transitionSorted(t *Transition) bool {
	upto := t.TransitionUpto
	if upto == r.states[2*t.Source] {
		// Transition isn't initialized yet (this is the first transition); don't check:
		return true
	}

	nextDest := r.transitions[upto]
	nextMin := r.transitions[upto+1]
	nextMax := r.transitions[upto+2]
	if nextMin > t.Min {
		return true
	} else if nextMin < t.Min {
		return false
	}

	// Min is equal, now test max:
	if nextMax > t.Max {
		return true
	} else if nextMax < t.Max {
		return false
	}

	// Max is also equal, now test dest:
	if nextDest > t.Dest {
		return true
	} else if nextDest < t.Dest {
		return false
	}

	// We should never see fully equal transitions here:
	return false
}

//Fill the provided Transition with the index'th transition leaving the specified state.
func (r *Automaton) getTransition(state, index int, t *Transition) {
	i := r.states[2*state] + 3*index
	t.Source = state
	t.Dest = r.transitions[i]
	i++
	t.Min = r.transitions[i]
	i++
	t.Max = r.transitions[i]
	i++
}

// Returns sorted array of all interval start points.
func (r *Automaton) getStartPoints() []int {
	pointset := make(map[int]struct{})
	pointset[0] = struct{}{}

	for s := 0; s < r.nextState; s += 2 {
		trans := r.states[s]
		limit := trans + 3*r.states[s+1]
		//System.out.println("  state=" + (s/2) + " trans=" + trans + " limit=" + limit);
		for trans < limit {
			min := r.transitions[trans+1]
			max := r.transitions[trans+2]
			//System.out.println("    min=" + min);
			pointset[min] = struct{}{}
			if max < 0x10FFFF {
				pointset[max+1] = struct{}{}
			}
			trans += 3
		}
	}

	points := make([]int, 0, len(pointset))
	for k, _ := range pointset {
		points = append(points, k)
	}
	sort.Ints(points)
	return points
}

// Step Performs lookup in transitions, assuming determinism.
// Params: 	state – starting state
//			label – codepoint to look up
// Returns: destination state, -1 if no matching outgoing transition
func (r *Automaton) Step(state, label int) int {
	return r.next(state, 0, label, nil)
}

// Next Looks for the next transition that matches the provided label, assuming determinism.
// This method is similar to step(int, int) but is used more efficiently when iterating over multiple
// transitions from the same source state. It keeps the latest reached transition index in
// transition.transitionUpto so the next call to this method can continue from there instead of restarting
// from the first transition.
// Params: 	transition – The transition to start the lookup from (inclusive, using its Transition.source
// 			and Transition.transitionUpto). It is updated with the matched transition; or with Transition.dest = -1 if no match.
// 			label – The codepoint to look up.
// Returns: The destination state; or -1 if no matching outgoing transition.
func (r *Automaton) Next(transition *Transition, label int) int {
	return r.next(transition.Source, 0, label, transition)
}

// Looks for the next transition that matches the provided label, assuming determinism.
// Params: 	state – The source state.
//			fromTransitionIndex – The transition index to start the lookup from (inclusive); negative interpreted as 0.
//			label – The codepoint to look up.
//			transition – The output transition to update with the matching transition; or null for no update.
// Returns: The destination state; or -1 if no matching outgoing transition.
func (r *Automaton) next(state, fromTransitionIndex, label int, transition *Transition) int {
	stateIndex := 2 * state
	firstTransitionIndex := r.states[stateIndex]
	numTransitions := r.states[stateIndex+1]

	// Since transitions are sorted,
	// binary search the transition for which label is within [minLabel, maxLabel].
	low := func() int {
		if fromTransitionIndex > 0 {
			return fromTransitionIndex
		}
		return 0
	}()
	high := numTransitions - 1

	for low <= high {
		mid := (low + high) >> 1
		transitionIndex := firstTransitionIndex + 3*mid
		minLabel := r.transitions[transitionIndex+1]
		if minLabel > label {
			high = mid - 1
		} else {
			maxLabel := r.transitions[transitionIndex+2]
			if maxLabel < label {
				low = mid + 1
			} else {
				destState := r.transitions[transitionIndex]
				if transition != nil {
					transition.Dest = destState
					transition.Min = minLabel
					transition.Max = maxLabel
					transition.TransitionUpto = mid
				}
				return destState
			}
		}
	}

	destState := -1
	if transition != nil {
		transition.Dest = destState
		transition.TransitionUpto = low
	}
	return destState
}

// AutomatonBuilder Records new states and transitions and then finish creates the Automaton. Use this
// when you cannot create the Automaton directly because it's too restrictive to have to add all transitions
// leaving each state at once.
type AutomatonBuilder struct {
	nextState      int
	isAccept       *bitset.BitSet
	transitions    []int
	nextTransition int
}

func (r *AutomatonBuilder) CreateState() int {
	res := r.nextState
	r.nextState++
	return res
}

func (r *AutomatonBuilder) SetAccept(state int, accept bool) {
	r.isAccept.SetTo(uint(state), accept)
}

// CopyStates Copies over all states from other.
func (r *AutomatonBuilder) CopyStates(other *Automaton) {
	otherNumStates := other.GetNumStates()
	for s := 0; s < otherNumStates; s++ {
		newState := r.CreateState()
		r.SetAccept(newState, other.IsAccept(s))
	}
}
