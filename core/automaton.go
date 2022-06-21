package core

import (
	"fmt"
	"github.com/bits-and-blooms/bitset"
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
	//stateOffset := r.GetNumStates()
	panic("")
}

// Freezes the last state, sorting and reducing the transitions.
func (r *Automaton) finishCurrentState() {
	//numTransitions := r.states[2*r.curState+1]
	//
	//offset := r.states[2*r.curState]
	//start := offset / 3
	panic("")
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
	panic("")
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
	panic("")
}

// Step Performs lookup in transitions, assuming determinism.
// Params: 	state – starting state
//			label – codepoint to look up
// Returns: destination state, -1 if no matching outgoing transition
func (r *Automaton) Step(state, label int) int {
	panic("")
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
	panic("")
}

// Looks for the next transition that matches the provided label, assuming determinism.
// Params: 	state – The source state.
//			fromTransitionIndex – The transition index to start the lookup from (inclusive); negative interpreted as 0.
//			label – The codepoint to look up.
//			transition – The output transition to update with the matching transition; or null for no update.
// Returns: The destination state; or -1 if no matching outgoing transition.
func (r *Automaton) next(state, fromTransitionIndex, label int, transition *Transition) int {
	panic("")
}
