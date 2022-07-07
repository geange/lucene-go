package index

import "github.com/geange/lucene-go/core/util"

// A Term represents a word from text. This is the unit of search. It is composed of two elements, the text of the
// word, as a string, and the name of the field that the text occurred in. Note that terms may represent more
// than words from text fields, but also things like dates, email addresses, urls, etc.
type Term struct {
	field string
	bytes *util.BytesRef
}
