package automaton

// Transition Holds one transition from an Automaton. This is typically used temporarily when iterating
// through transitions by invoking Automaton.initTransition and Automaton.getNextTransition.
type Transition struct {
	// Source state.
	Source int

	// Destination state.
	Dest int

	// Minimum accepted label (inclusive).
	Min int

	// Maximum accepted label (inclusive).
	Max int

	// Remembers where we are in the iteration; init to -1 to provoke exception if nextTransition is
	// called without first initTransition.
	TransitionUpto int
}

func NewTransition() *Transition {
	return &Transition{}
}
