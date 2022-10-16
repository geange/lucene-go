package fst

import "github.com/geange/lucene-go/core/store"

type FixedLengthArcsBuffer struct {
	// Initial capacity is the max length required for the header of a node with fixed length arcs:
	// header(byte) + numArcs(vint) + numBytes(vint)
	bytes []byte

	bado *store.ByteArrayDataOutput
}
