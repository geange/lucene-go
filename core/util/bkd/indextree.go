package bkd

import (
	"context"
	"github.com/geange/lucene-go/core/store"
	"io"
	"slices"
)

// IndexTree
// Used to walk the off-heap index. The format takes advantage of the limited access pattern to
// the BKD tree at search time, i.e. starting at the root node and recursing downwards one child at a time.
// 用于遍历堆外索引。该格式利用了在搜索时对BKD树的有限访问模式，即从根节点开始，一次向下递归一个子节点。
// lucene.internal
type IndexTree struct {
	reader *Reader

	nodeID int

	// level is 1-based so that we can do level-1 w/o checking each time:
	level int

	splitDim int

	splitPackedValueStack [][]byte

	// used to read the packed tree off-heap
	in store.IndexInput

	// holds the minimum (left most) leaf block file pointer for each level we've recursed to:
	leafBlockFPStack []int64

	// holds the address, in the off-heap index, of the right-node of each level:
	rightNodePositions []int

	// holds the splitDim for each level:
	splitDims []int

	// true if the per-dim delta we read for the node at this level is a negative offset vs. the last split on this dim; this is a packed
	// 2D array, i.e. to access array[level][dim] you read from negativeDeltas[level*numDims+dim].  this will be true if the last time we
	// split on this dimension, we next pushed to the left sub-tree:
	negativeDeltas []bool

	// holds the packed per-level split values; the intersect method uses this to save the cell min/max as it recurses:
	splitValuesStack [][]byte

	// scratch value to return from getPackedValue:
	scratch []byte
}

func (r *Reader) newIndexTree(ctx context.Context) (*IndexTree, error) {
	in := r.packedIndex.Clone().(store.IndexInput)
	index := r.newIndexTreeV1(in, 1, 1)
	err := index.readNodeData(ctx, false)
	if err != nil {
		return nil, err
	}
	return index, nil
}

func (r *Reader) newIndexTreeV1(in store.IndexInput, nodeID int, level int) *IndexTree {
	treeDepth := r.getTreeDepth()

	splitPackedValueStack := make([][]byte, treeDepth+1)
	splitPackedValueStack[level] = make([]byte, r.config.PackedIndexBytesLength())

	splitValuesStack := make([][]byte, treeDepth+1)
	splitValuesStack[0] = make([]byte, r.config.packedIndexBytesLength)

	return &IndexTree{
		reader:                r,
		nodeID:                nodeID,
		level:                 level,
		splitDim:              0,
		splitPackedValueStack: splitPackedValueStack,
		in:                    in,
		leafBlockFPStack:      make([]int64, treeDepth+1),
		rightNodePositions:    make([]int, treeDepth+1),
		splitDims:             make([]int, treeDepth+1),
		negativeDeltas:        make([]bool, r.config.NumIndexDims()*(treeDepth+1)),
		splitValuesStack:      splitValuesStack,
		scratch:               make([]byte, r.config.BytesPerDim()),
	}
}

func (t *IndexTree) PushLeft(ctx context.Context) error {
	t.nodeID *= 2
	t.level++
	return t.readNodeData(ctx, true)
}

func (t *IndexTree) Clone() *IndexTree {
	clone := t.in.Clone().(store.IndexInput)
	config := t.reader.config

	index := t.reader.newIndexTreeV1(clone, t.nodeID, t.level)
	// copy node data
	index.splitDim = t.splitDim
	index.leafBlockFPStack[t.level] = t.leafBlockFPStack[t.level]
	index.rightNodePositions[t.level] = t.rightNodePositions[t.level]
	index.splitValuesStack[index.level] = slices.Clone(t.splitValuesStack[index.level])
	arraycopy(t.negativeDeltas, t.level*config.numIndexDims, index.negativeDeltas, t.level*config.numIndexDims, config.numIndexDims)
	index.splitDims[t.level] = t.splitDims[t.level]
	return index
}

func (t *IndexTree) PushRight(ctx context.Context) error {
	nodePosition := t.rightNodePositions[t.level]
	t.nodeID = t.nodeID*2 + 1
	t.level++

	_, err := t.in.Seek(int64(nodePosition), io.SeekStart)
	if err != nil {
		return err
	}
	return t.readNodeData(ctx, false)
}

func (t *IndexTree) Pop() {
	t.nodeID /= 2
	t.level--
	t.splitDim = t.splitDims[t.level]
}

func (t *IndexTree) IsLeafNode() bool {
	return t.nodeID >= t.reader.leafNodeOffset
}

func (t *IndexTree) NodeExists() bool {
	return t.nodeID-t.reader.leafNodeOffset < t.reader.leafNodeOffset
}

func (t *IndexTree) GetNodeID() int {
	return t.nodeID
}

func (t *IndexTree) GetSplitPackedValue() []byte {
	return t.splitPackedValueStack[t.level]
}

// GetSplitDim Only valid after pushLeft or pushRight, not pop!
func (t *IndexTree) GetSplitDim() int {
	return t.splitDim
}

// GetSplitDimValue Only valid after pushLeft or pushRight, not pop!
func (t *IndexTree) GetSplitDimValue() []byte {
	config := t.reader.config
	t.scratch = t.splitValuesStack[t.level]
	offset := t.splitDim * config.bytesPerDim
	size := config.bytesPerDim
	return t.scratch[offset : offset+size]
}

// GetLeafBlockFP Only valid after pushLeft or pushRight, not pop!
func (t *IndexTree) GetLeafBlockFP() int64 {
	return t.leafBlockFPStack[t.level]
}

// Return the number of leaves below the current node.
// 获取当前节点的叶子数
func (t *IndexTree) getNumLeaves() int {
	leftMostLeafNode := t.nodeID
	for leftMostLeafNode < t.reader.leafNodeOffset {
		leftMostLeafNode = leftMostLeafNode * 2
	}

	rightMostLeafNode := t.nodeID
	for rightMostLeafNode < t.reader.leafNodeOffset {
		rightMostLeafNode = rightMostLeafNode*2 + 1
	}

	var numLeaves int
	if rightMostLeafNode >= leftMostLeafNode {
		// both are on the same level
		numLeaves = rightMostLeafNode - leftMostLeafNode + 1
	} else {
		// left is one level deeper than right
		numLeaves = rightMostLeafNode - leftMostLeafNode + 1 + t.reader.leafNodeOffset
	}
	return numLeaves
}

func (t *IndexTree) readNodeData(ctx context.Context, isLeft bool) error {
	config := t.reader.config

	level := t.level

	if t.splitPackedValueStack[level] == nil {
		t.splitPackedValueStack[level] = make([]byte, config.packedIndexBytesLength)
	}
	arraycopy(t.negativeDeltas, (level-1)*config.numIndexDims, t.negativeDeltas, level*config.numIndexDims, config.numIndexDims)
	//assert splitDim != -1;
	t.negativeDeltas[level*config.numIndexDims+t.splitDim] = isLeft

	t.leafBlockFPStack[level] = t.leafBlockFPStack[level-1]

	// read leaf block FP delta
	if isLeft == false {
		n, err := t.in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		t.leafBlockFPStack[level] += int64(n)
	}

	if t.IsLeafNode() {
		t.splitDim = -1
	} else {

		// read split dim, prefix, firstDiffByteDelta encoded as int:
		n, err := t.in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		code := int(n)
		t.splitDim = code % config.numIndexDims
		t.splitDims[level] = t.splitDim
		code /= config.numIndexDims
		prefix := code % (1 + config.bytesPerDim)
		suffix := config.bytesPerDim - prefix

		if t.splitValuesStack[level] == nil {
			t.splitValuesStack[level] = make([]byte, config.packedIndexBytesLength)
		}
		arraycopy(t.splitValuesStack[level-1], 0, t.splitValuesStack[level], 0, config.packedIndexBytesLength)
		if suffix > 0 {
			firstDiffByteDelta := code / (1 + config.bytesPerDim)
			if t.negativeDeltas[level*config.numIndexDims+t.splitDim] {
				firstDiffByteDelta = -firstDiffByteDelta
			}
			oldByte := t.splitValuesStack[level][t.splitDim*config.bytesPerDim+prefix]
			t.splitValuesStack[level][t.splitDim*config.bytesPerDim+prefix] = (byte)(oldByte + byte(firstDiffByteDelta))

			offset := t.splitDim*config.bytesPerDim + prefix + 1
			size := suffix - 1
			t.in.Read(t.splitValuesStack[level][offset : offset+size])
		} else {
			// our split value is == last split value in this dim, which can happen when there are many duplicate values
		}

		var leftNumBytes int
		if t.nodeID*2 < t.reader.leafNodeOffset {
			n, err := t.in.ReadUvarint(ctx)
			if err != nil {
				return err
			}
			leftNumBytes = int(n)
		} else {
			leftNumBytes = 0
		}
		t.rightNodePositions[level] = int(t.in.GetFilePointer()) + leftNumBytes
	}

	return nil

	//numIndexDims := config.NumIndexDims()
	//bytesPerDim := config.BytesPerDim()
	//
	//if t.splitPackedValueStack[t.level] == nil {
	//	t.splitPackedValueStack[t.level] = make([]byte, config.packedIndexBytesLength)
	//}
	//arraycopy(t.negativeDeltas, (t.level-1)*config.numIndexDims, t.negativeDeltas, t.level*config.numIndexDims, config.numIndexDims)
	//
	//t.negativeDeltas[t.level*numIndexDims+t.splitDim] = isLeft
	//
	//t.leafBlockFPStack[t.level] = t.leafBlockFPStack[t.level-1]
	//
	//// read leaf block FP delta
	//if isLeft == false {
	//	num, err := t.in.ReadUvarint()
	//	if err != nil {
	//		return err
	//	}
	//	t.leafBlockFPStack[t.level] += int64(num)
	//}
	//
	//if t.IsLeafNode() {
	//	t.splitDim = -1
	//	return nil
	//}
	//
	//// read split dim, prefix, firstDiffByteDelta encoded as int:
	//num, err := t.in.ReadUvarint()
	//if err != nil {
	//	return err
	//}
	//code := int(num)
	//t.splitDim = code % numIndexDims
	//t.splitDims[t.level] = t.splitDim
	//code /= numIndexDims
	//prefix := code % (1 + bytesPerDim)
	//suffix := bytesPerDim - prefix
	//
	//if t.splitValuesStack[t.level] == nil {
	//	t.splitValuesStack[t.level] = make([]byte, config.packedIndexBytesLength)
	//}
	//
	////copySize := config.packedIndexBytesLength
	//arraycopy(t.splitValuesStack[t.level-1], 0, t.splitValuesStack[t.level], 0, config.packedIndexBytesLength)
	//if suffix > 0 {
	//	firstDiffByteDelta := code / (1 + bytesPerDim)
	//	if t.negativeDeltas[t.level*numIndexDims+t.splitDim] {
	//		firstDiffByteDelta = -firstDiffByteDelta
	//	}
	//	oldByte := t.splitValuesStack[t.level][t.splitDim*bytesPerDim+prefix]
	//	t.splitValuesStack[t.level][t.splitDim*bytesPerDim+prefix] = oldByte + byte(firstDiffByteDelta)
	//
	//	offset := t.splitDim*bytesPerDim + prefix + 1
	//	size := suffix - 1
	//	_, err := t.in.Read(t.splitValuesStack[t.level][offset : offset+size])
	//	if err != nil {
	//		return err
	//	}
	//} else {
	//	// our split value is == last split value in this dim, which can happen when there are many duplicate values
	//}
	//
	//var leftNumBytes int
	//if t.nodeID*2 < t.reader.leafNodeOffset {
	//	variant, err := t.in.ReadUvarint()
	//	if err != nil {
	//		return err
	//	}
	//	leftNumBytes = int(variant)
	//} else {
	//	leftNumBytes = 0
	//}
	//
	//fp := t.in.GetFilePointer()
	//
	//t.rightNodePositions[t.level] = int(fp) + leftNumBytes
	//return nil
}
