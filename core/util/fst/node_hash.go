package fst

import "github.com/geange/lucene-go/core/util/packed"

type NodeHash struct {
	table      *packed.PagedGrowableWriter
	count      int64
	mask       int64
	fst        *FST
	scratchArc *FSTArc
	in         BytesReader
}

func NewNodeHash(table *packed.PagedGrowableWriter) *NodeHash {
	return &NodeHash{table: table}
}


func (n *NodeHash) nodesEqual(node *UnCompiledNode, address int64) (bool, error) {
	panic("")
}
