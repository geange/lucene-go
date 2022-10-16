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
