package index

// TermStates Maintains a IndexReader TermState view over IndexReader instances containing a single term.
// The TermStates doesn't track if the given TermState objects are valid, neither if the TermState instances
// refer to the same terms in the associated readers.
type TermStates struct {
	topReaderContextIdentity any
	states                   []TermState
	term                     Term
	docFreq                  int
	totalTermFreq            int64
}
