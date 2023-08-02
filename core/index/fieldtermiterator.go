package index

import "github.com/geange/lucene-go/core/util/bytesref"

// FieldTermIterator
// Iterates over terms in across multiple fields.
// The caller must check field after each next to see if the field changed,
// but == can be used since the iterator implementation ensures it will use the same String instance for a given field.
type FieldTermIterator interface {
	bytesref.BytesIterator

	// Field
	// Returns current field.
	// This method should not be called after iteration is done.
	// Note that you may use == to detect a change in field.
	Field() string

	// DelGen Del gen of the current term.
	DelGen() int64
}
