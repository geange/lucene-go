package automaton

import "github.com/bits-and-blooms/bitset"

// Builder Records new states and transitions and then finish creates the Automaton. Use this
// when you cannot create the Automaton directly because it's too restrictive to have to add all transitions
// leaving each state at once.
type Builder struct {
	nextState      int
	isAccept       *bitset.BitSet
	transitions    []int
	nextTransition int
}

func NewNewBuilder() *Builder {
	return NewBuilderV1(16, 16)
}

func NewBuilderV1(numStates, numTransitions int) *Builder {
	return &Builder{
		nextState:      0,
		isAccept:       bitset.New(uint(numStates)),
		transitions:    make([]int, 4*numTransitions),
		nextTransition: 0,
	}
}

func (r *Builder) CreateState() int {
	res := r.nextState
	r.nextState++
	return res
}

func (r *Builder) SetAccept(state int, accept bool) {
	r.isAccept.SetTo(uint(state), accept)
}

// CopyStates Copies over all states from other.
func (r *Builder) CopyStates(other *Automaton) {
	otherNumStates := other.GetNumStates()
	for s := 0; s < otherNumStates; s++ {
		newState := r.CreateState()
		r.SetAccept(newState, other.IsAccept(s))
	}
}

func (r *Builder) AddTransitionLabel(source, dest, label int) {
	r.AddTransition(source, dest, label, label)
}

func (r *Builder) AddTransition(source, dest, min, max int) {
	if len(r.transitions) < r.nextTransition+4 {
		r.transitions = append(r.transitions, make([]int, 4)...)
	}
	r.transitions[r.nextTransition] = source
	r.nextTransition++
	r.transitions[r.nextTransition] = dest
	r.nextTransition++
	r.transitions[r.nextTransition] = min
	r.nextTransition++
	r.transitions[r.nextTransition] = max
	r.nextTransition++
}

func (r *Builder) Finish() *Automaton {
	// Create automaton with the correct size.
	numStates := r.nextState
	numTransitions := r.nextTransition / 4
	a := NewAutomatonV1(numStates, numTransitions)

	// Create all states.
	for state := 0; state < numStates; state++ {
		a.CreateState()
		a.SetAccept(state, r.IsAccept(state))
	}

	// Create all transitions
	r.sort(0, numTransitions)
	for upto := 0; upto < r.nextTransition; upto += 4 {
		a.AddTransition(r.transitions[upto],
			r.transitions[upto+1],
			r.transitions[upto+2],
			r.transitions[upto+3])
	}

	a.finishState()

	return a
}
