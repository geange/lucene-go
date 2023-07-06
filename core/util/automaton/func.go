package automaton

func MakeAnyString() *Automaton {
	a := NewAutomaton()
	s := a.CreateState()
	a.SetAccept(s, true)
	a.AddTransition(s, s, 0, 255)
	a.finishState()
	return a
}
