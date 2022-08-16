package fst

// NodeHash Used to dedup states (lookup already-frozen states)
type NodeHash[T any] struct {
	count      int64
	mask       int64
	fst        *FST[T]
	scratchArc *Arc[T]
	in         BytesReader
}

func (n *NodeHash[T]) AddNew(address int64, nodeIn *UnCompiledNode[T]) error {
	panic("")
}

// hash code for an unfrozen node.  This must be identical
// to the frozen case (below)!!
func (n *NodeHash[T]) hash(node int64) (int64, error) {
	panic("")
}
