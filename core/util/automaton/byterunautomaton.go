package automaton

// ByteRunAutomaton Automaton representation for matching UTF-8 byte[].
type ByteRunAutomaton struct {
	*RunAutomaton
}

func NewByteRunAutomaton(a *Automaton, isBinary bool, determinizeWorkLimit int) *ByteRunAutomaton {
	var auto *Automaton

	if isBinary {
		auto = a
	} else {

	}

	return &ByteRunAutomaton{
		NewRunAutomaton(auto, 256, determinizeWorkLimit),
	}
}

func (a *Automaton) NewByteRunAutomaton() *ByteRunAutomaton {
	return &ByteRunAutomaton{
		NewRunAutomaton(a, 256, 10000),
	}
}

// Run Returns true if the given byte array is accepted by this automaton
func (r *ByteRunAutomaton) Run(s []byte) bool {
	p := 0
	for i := 0; i < len(s); i++ {
		p = r.Step(p, int(s[i]&0xFF))
		if p == -1 {
			return false
		}
	}
	return r.accept[p]
}
