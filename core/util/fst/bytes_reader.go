package fst

import "github.com/geange/lucene-go/core/store"

// BytesReader Reads bytes stored in an FST.
type BytesReader interface {
	store.DataInput

	GetPosition() int64

	SetPosition(pos int64)

	// Reversed Returns true if this reader uses reversed bytes under-the-hood.
	Reversed() bool
}
