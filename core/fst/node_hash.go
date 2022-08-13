package fst

// NodeHash Used to dedup states (lookup already-frozen states)
type NodeHash struct {
	count      int64
	mask       int64
	fst        *FST
	scratchArc *Arc
	in         BytesReader
}
