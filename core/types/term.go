package types

import (
	"bytes"
	"strings"
)

// A Term represents a word from text. This is the unit of search. It is composed of two elements, the text of the
// word, as a string, and the name of the field that the text occurred in. Note that terms may represent more
// than words from text fields, but also things like dates, email addresses, urls, etc.
type Term struct {
	field  string
	values []byte
}

func NewTerm(field string, values []byte) *Term {
	return &Term{field: field, values: values}
}

// Field
// Returns the field of this term. The field indicates the part of a document which this term came from.
func (r *Term) Field() string {
	return r.field
}

// Text Returns the text of this term. In the case of words, this is simply the text of the word. In the case
// of dates and other types, this is an encoding of the object as a string.
func (r *Term) Text() string {
	return string(r.values)
}

func (r *Term) Bytes() []byte {
	return r.values
}

func TermCompare(a, b *Term) int {
	cmp := strings.Compare(a.field, b.field)
	if cmp != 0 {
		return cmp
	}
	return bytes.Compare(a.values, b.values)
}

// TermState Encapsulates all required internal state to position the associated TermsEnum without re-seeking.
// See Also: TermsEnum.seekExact(org.apache.lucene.util.BytesRef, TermState), TermsEnum.termState()
type TermState interface {

	// CopyFrom
	// Copies the content of the given TermState to this instance
	// other – the TermState to copy
	CopyFrom(other TermState)
}

var _ TermState = &OrdTermState{}

type OrdTermState struct {
	Ord int64
}

func NewOrdTermState() *OrdTermState {
	return &OrdTermState{}
}

func (r *OrdTermState) CopyFrom(other TermState) {
	if v, ok := other.(*OrdTermState); ok {
		r.Ord = v.Ord
	}
}
