package core

// CompiledAutomaton Immutable class holding compiled details for a given Automaton. The Automaton is
// deterministic, must not have dead states but is not necessarily minimal.
type CompiledAutomaton struct {
	// If simplify is true this will be the "simplified" type; else, this is NORMAL
	_type int

	// For CompiledAutomaton.AUTOMATON_TYPE.SINGLE this is the singleton term.
	term []byte

	// Matcher for quickly determining if a byte[] is accepted. only valid for CompiledAutomaton.AUTOMATON_TYPE.NORMAL.
	runAutomaton *ByteRunAutomaton

	// Two dimensional array of transitions, indexed by state number for traversal. The state numbering is
	// consistent with runAutomaton. Only valid for CompiledAutomaton.AUTOMATON_TYPE.NORMAL.
	automaton *Automaton

	// Shared common suffix accepted by the automaton. Only valid for CompiledAutomaton.AUTOMATON_TYPE.NORMAL,
	// and only when the automaton accepts an infinite language. This will be null if the common prefix is length 0.
	commonSuffixRef []byte

	// Indicates if the automaton accepts a finite set of strings. Null if this was not computed. Only valid
	// for CompiledAutomaton.AUTOMATON_TYPE.NORMAL.
	finite bool

	// Which state, if any, accepts all suffixes, else -1.
	sinkState bool

	transition *Transition
}

const (
	AUTOMATON_TYPE_NONE   = iota // Automaton that accepts no strings.
	AUTOMATON_TYPE_ALL           // Automaton that accepts all possible strings.
	AUTOMATON_TYPE_SINGLE        // Automaton that accepts only a single fixed string.
	AUTOMATON_TYPE_NORMAL        // Catch-all for any other automata.
)
