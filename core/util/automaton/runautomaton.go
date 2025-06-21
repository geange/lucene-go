package automaton

// RunAutomaton Finite-state automaton with fast run operation. The initial state is always 0.
type RunAutomaton struct {
	automaton    *Automaton
	alphabetSize int
	size         int
	accept       []bool

	// delta(state,c) = transitions[state*points.length +
	// getCharClass(c)]
	transitions []int

	// char interval start points
	points []int

	// map from char number to class
	classmap []int
}

func NewRunAutomaton(a *Automaton, alphabetSize, determinizeWorkLimit int) *RunAutomaton {
	size := max(1, a.GetNumStates())
	points := a.GetStartPoints()

	r := RunAutomaton{
		automaton:    DeterminizeAutomaton(a, determinizeWorkLimit),
		alphabetSize: alphabetSize,
		size:         size,
		accept:       make([]bool, size),
		transitions:  make([]int, size*len(points)),
		points:       points,
		classmap:     make([]int, min(256, alphabetSize)),
	}

	for i := 0; i < len(r.transitions); i++ {
		r.transitions[i] = -1
	}

	transition := &Transition{}

	for n := 0; n < size; n++ {
		r.accept[n] = a.IsAccept(n)
		transition.Source = n
		transition.TransitionUpto = -1
		for c := 0; c < len(r.points); c++ {
			dest := a.Next(transition, r.points[c])
			r.transitions[n*len(r.points)+c] = dest
		}
	}

	i := 0
	for j := 0; j < len(r.classmap); j++ {
		if i+1 < len(r.points) && j == points[i+1] {
			i++
		}
		r.classmap[j] = i
	}

	return &r
}

// GetSize Returns number of states in automaton.
func (r *RunAutomaton) GetSize() int {
	return r.size
}

// IsAccept Returns acceptance status for given state.
func (r *RunAutomaton) IsAccept(state int) bool {
	return r.accept[state]
}

// Returns array of codepoint class interval start points. The array should not be modified by the caller.
func (r *RunAutomaton) getCharIntervals() []int {
	res := make([]int, len(r.points))
	copy(res, r.points)
	return res
}

// GetCharClass Gets character class of given codepoint
func (r *RunAutomaton) GetCharClass(c int) int {
	// binary search
	a := 0
	b := len(r.points)
	for b-a > 1 {
		d := (a + b) >> 1
		if r.points[d] > c {
			b = d
		} else if r.points[d] < c {
			a = d
		} else {
			return d
		}
	}
	return a
}

// Step Returns the state obtained by reading the given char from the given state. Returns -1 if not obtaining
// any such state. (If the original Automaton had no dead states, -1 is returned here if and only if a dead
// state is entered in an equivalent automaton with a total transition function.)
func (r *RunAutomaton) Step(state int, c int) int {
	if c >= len(r.classmap) {
		return r.transitions[state*len(r.points)+r.GetCharClass(c)]
	}
	return r.transitions[state*len(r.points)+r.classmap[c]]
}
