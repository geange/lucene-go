package automaton

func Run(a *Automaton, s string) bool {
	state := 0
	for _, v := range s {
		nextState := a.Step(state, int(v))
		if nextState == -1 {
			return false
		}
		state = nextState
	}
	return a.IsAccept(state)
}
