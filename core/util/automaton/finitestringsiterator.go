package automaton

// FiniteStringsIterator
// Iterates all accepted strings.
// If the Automaton has cycles then this iterator may throw an IllegalArgumentException, but this is not guaranteed!
// Be aware that the iteration order is implementation dependent and may change across releases.
// If the automaton is not determinized then it's possible this iterator will return duplicates.
type FiniteStringsIterator struct {
}

func NewFiniteStringsIterator(a *Automaton, startState, endState int) *FiniteStringsIterator {
	panic("")
}

type PathNode struct {
	state      int
	to         int
	transition int
	label      int
	t          *Transition
}
