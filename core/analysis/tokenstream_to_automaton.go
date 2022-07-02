package analysis

import (
	"github.com/geange/lucene-go/core/tokenattributes"
	"github.com/geange/lucene-go/core/util/automaton"
)

// TokenStreamToAutomaton Consumes a TokenStream and creates an Automaton where the transition labels are UTF8
// bytes (or Unicode code points if unicodeArcs is true) from the TermToBytesRefAttribute. Between tokens we
// insert POS_SEP and for holes we insert HOLE.
type TokenStreamToAutomaton struct {
	preservePositionIncrements bool
	finalOffsetGapAsHole       bool
	unicodeArcs                bool
}

func NewTokenStreamToAutomaton() *TokenStreamToAutomaton {
	return &TokenStreamToAutomaton{preservePositionIncrements: true}
}

// SetPreservePositionIncrements Whether to generate holes in the automaton for missing positions, true by default.
func (r *TokenStreamToAutomaton) SetPreservePositionIncrements(enablePositionIncrements bool) {
	r.preservePositionIncrements = enablePositionIncrements
}

// SetFinalOffsetGapAsHole f true, any final offset gaps will result in adding a position hole.
func (r *TokenStreamToAutomaton) SetFinalOffsetGapAsHole(finalOffsetGapAsHole bool) {
	r.finalOffsetGapAsHole = finalOffsetGapAsHole
}

// SetUnicodeArcs Whether to make transition labels Unicode code points instead of UTF8 bytes, false by default
func (r *TokenStreamToAutomaton) SetUnicodeArcs(unicodeArcs bool) {
	r.unicodeArcs = unicodeArcs
}

// ChangeToken Subclass and implement this if you need to change the token (such as escaping certain bytes)
// before it's turned into a graph.
func (r *TokenStreamToAutomaton) ChangeToken(in []byte) []byte {
	return in
}

const (
	// POS_SEP We create transition between two adjacent tokens.
	POS_SEP = 0x001f

	// HOLE We add this arc to represent a hole.
	HOLE = 0x001e
)

func (r *TokenStreamToAutomaton) ToAutomaton(in TokenStream) (*automaton.Automaton, error) {
	builder := automaton.NewNewBuilderDefault()
	builder.CreateState()

	in.GetAttributeSource().Add(tokenattributes.NewPackedTokenAttributeImpl())

	panic("")
}
