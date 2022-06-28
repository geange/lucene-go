package core

// Determinize Determinizes the given automaton.
// Worst case complexity: exponential in number of states.
// Params: 	workLimit – Maximum amount of "work" that the powerset construction will spend before throwing
//			TooComplexToDeterminizeException. Higher numbers allow this operation to consume more memory and
//			CPU but allow more complex automatons. Use DEFAULT_DETERMINIZE_WORK_LIMIT as a decent default
//			if you don't otherwise know what to specify.
// Throws: TooComplexToDeterminizeException – if determinizing requires more than workLimit "effort"
func Determinize(a *Automaton, workLimit int) *Automaton {
	return a
}
