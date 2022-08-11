package index

import "github.com/emirpasic/gods/utils"

// A Term represents a word from text. This is the unit of search. It is composed of two elements, the text of the
// word, as a string, and the name of the field that the text occurred in. Note that terms may represent more
// than words from text fields, but also things like dates, email addresses, urls, etc.
type Term struct {
	field string
	bytes []byte
}

func NewTerm(field string, bytes []byte) *Term {
	return &Term{field: field, bytes: bytes}
}

// Field Returns the field of this term. The field indicates the part of a document which this term came from.
func (r *Term) Field() string {
	return r.field
}

// Text Returns the text of this term. In the case of words, this is simply the text of the word. In the case
// of dates and other types, this is an encoding of the object as a string.
func (r *Term) Text() string {
	return string(r.bytes)
}

func (r *Term) Bytes() []byte {
	return r.bytes
}

func TermComparator(a, b any) int {
	c1, c2 := a.(Term), b.(Term)
	cmp := utils.StringComparator(c1.field, c2.field)
	if cmp == 0 {
		return utils.ByteComparator(c1.bytes, c2.bytes)
	}
	return cmp
}
