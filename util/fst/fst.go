package fst

import "github.com/geange/lucene-go/core/store"

// BytesReader Reads bytes stored in an FST.
type BytesReader interface {
	store.DataInput

	// GetPosition Get current read position.
	GetPosition() int

	// SetPosition Set current read position.
	SetPosition(pos int)

	// Reversed Returns true if this reader uses reversed bytes under-the-hood.
	Reversed() bool
}