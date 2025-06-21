package automaton

// Minimize
// Minimizes (and determinizes if not already deterministic) the given automaton using Hopcroft's algorithm.
func Minimize(a *Automaton, determinizeWorkLimit int) (*Automaton, error) {
	if a.GetNumStates() == 0 || (a.IsAccept(0) == false && a.GetNumTransitionsWithState(0) == 0) {
		// Fastmatch for common case
		return NewAutomaton(), nil
	}

	// TODO: fix it
	return determinize(a, determinizeWorkLimit)
}

type IntPair struct {
	n1 int
	n2 int
}

type StateList struct {
}

type StateListNode struct {
}
