package bkd

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

// IntersectState Used to track all state for a single call to intersect.
// 用于跟踪要相交的单个调用的所有状态。
type IntersectState struct {
	in                         store.IndexInput
	scratchIterator            *readerDocIDSetIterator
	scratchDataPackedValue     []byte
	scratchMinIndexPackedValue []byte
	scratchMaxIndexPackedValue []byte
	commonPrefixLengths        []int
	visitor                    types.IntersectVisitor
	index                      *IndexTree
}

func NewIntersectState(in store.IndexInput, config *Config,
	visitor types.IntersectVisitor, indexVisitor *IndexTree) *IntersectState {

	return &IntersectState{
		in:                         in,
		scratchIterator:            newReaderDocIDSetIterator(config.maxPointsInLeafNode),
		scratchDataPackedValue:     make([]byte, config.packedBytesLength),
		scratchMinIndexPackedValue: make([]byte, config.packedIndexBytesLength),
		scratchMaxIndexPackedValue: make([]byte, config.packedIndexBytesLength),
		commonPrefixLengths:        make([]int, config.numDims),
		visitor:                    visitor,
		index:                      indexVisitor,
	}
}
