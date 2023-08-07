package bkd

import (
	"bytes"
	"errors"
	"io"
	"slices"

	linked "github.com/geange/gods-generic/lists/singlylinkedlist"
	"github.com/geange/lucene-go/core/store"
)

// Packs the two arrays, representing a semi-balanced binary tree, into a compact byte[] structure.
// 将表示半平衡二进制树的两个数组打包为紧凑的byte[]结构。
func (w *Writer) packIndex(leafNodes LeafNodes) ([]byte, error) {
	config := w.config

	bytesPerDim := config.BytesPerDim()
	numIndexDims := config.NumIndexDims()

	// Reused while packing the index
	writeBuffer := store.NewByteBuffersDataOutput()

	blocks := linked.NewWith[[]byte](bytes.Compare)

	lastSplitValues := make([]byte, bytesPerDim*numIndexDims)

	totalSize, err := w.recursePackIndex(writeBuffer, leafNodes, 0, blocks, lastSplitValues,
		make([]bool, numIndexDims), false, 0, leafNodes.NumLeaves())
	if err != nil {
		return nil, err
	}

	// Compact the byte[] blocks into single byte index:
	index := make([]byte, totalSize)
	upto := 0

	for _, block := range blocks.Values() {
		arraycopy(block, 0, index, upto, len(block))
		upto += len(block)
	}
	return index, nil
}

// Appends the current contents of writeBuffer as another block on the growing in-memory file
func (w *Writer) appendBlock(writeBuffer *store.ByteBuffersDataOutput, blocks *linked.List[[]byte]) int {
	block := slices.Clone(writeBuffer.Bytes())
	blocks.Add(block)
	writeBuffer.Reset()
	return len(block)
}

// lastSplitValues is per-dimension split value previously seen;
// we use this to prefix-code the split byte[] on each inner node
func (w *Writer) recursePackIndex(writeBuffer *store.ByteBuffersDataOutput, leafNodes LeafNodes,
	minBlockFP int64, blocks *linked.List[[]byte], lastSplitValues []byte, negativeDeltas []bool,
	isLeft bool, leavesOffset, numLeaves int) (int, error) {

	config := w.config
	if numLeaves == 1 {
		if isLeft {
			//assert leafNodes.getLeafLP(leavesOffset) - minBlockFP == 0;
			return 0, nil
		} else {
			delta := leafNodes.GetLeafLP(leavesOffset) - minBlockFP
			//assert leafNodes.numLeaves() == numLeaves || delta > 0 : "expected delta > 0; got numLeaves =" + numLeaves + " and delta=" + delta;
			if err := writeBuffer.WriteUvarint(uint64(delta)); err != nil {
				return 0, err
			}
			return w.appendBlock(writeBuffer, blocks), nil
		}
	}

	var leftBlockFP int64
	if isLeft {
		// The left tree's left most leaf block FP is always the minimal FP:
		//assert leafNodes.getLeafLP(leavesOffset) == minBlockFP;
		leftBlockFP = minBlockFP
	} else {
		leftBlockFP = leafNodes.GetLeafLP(leavesOffset)
		delta := leftBlockFP - minBlockFP
		//assert leafNodes.numLeaves() == numLeaves || delta > 0 : "expected delta > 0; got numLeaves =" + numLeaves + " and delta=" + delta;
		if err := writeBuffer.WriteUvarint(uint64(delta)); err != nil {
			return 0, err
		}
	}

	numLeftLeafNodes := w.getNumLeftLeafNodes(numLeaves)
	rightOffset := leavesOffset + numLeftLeafNodes
	splitOffset := rightOffset - 1

	splitDim := leafNodes.GetSplitDimension(splitOffset)
	splitValue := leafNodes.GetSplitValue(splitOffset)
	address := 0

	// find common prefix with last split value in this dim:
	prefix := Mismatch(splitValue[address:address+config.bytesPerDim],
		lastSplitValues[splitDim*config.bytesPerDim:splitDim*config.bytesPerDim+config.bytesPerDim])
	if prefix == -1 {
		prefix = config.bytesPerDim
	}

	var firstDiffByteDelta int
	if prefix < config.bytesPerDim {
		firstDiffByteDelta = int(splitValue[address+prefix]) - int(lastSplitValues[splitDim*config.bytesPerDim+prefix])
		if negativeDeltas[splitDim] {
			firstDiffByteDelta = -firstDiffByteDelta
		}
		//assert firstDiffByteDelta > 0;
	} else {
		firstDiffByteDelta = 0
	}

	// pack the prefix, splitDim and delta first diff byte into a single vInt:
	code := (firstDiffByteDelta*(1+config.bytesPerDim)+prefix)*config.numIndexDims + splitDim

	if err := writeBuffer.WriteUvarint(uint64(code)); err != nil {
		return 0, err
	}

	// write the split value, prefix coded vs. our parent's split value:
	suffix := config.bytesPerDim - prefix
	savSplitValue := make([]byte, suffix)
	if suffix > 1 {
		idx := address + prefix + 1
		size := suffix - 1
		if _, err := writeBuffer.Write(splitValue[idx : idx+size]); err != nil {
			return 0, err
		}
	}

	//cmp := slices.Clone(lastSplitValues)

	arraycopy(lastSplitValues, splitDim*config.bytesPerDim+prefix, savSplitValue, 0, suffix)

	// copy our split value into lastSplitValues for our children to prefix-code against
	arraycopy(splitValue, address+prefix, lastSplitValues, splitDim*config.bytesPerDim+prefix, suffix)

	numBytes := w.appendBlock(writeBuffer, blocks)

	// placeholder for left-tree numBytes; we need this so that at search time if we only need to recurse into the right sub-tree we can
	// quickly seek to its starting point
	idxSav := blocks.Size()
	blocks.Add(nil)

	savNegativeDelta := negativeDeltas[splitDim]
	negativeDeltas[splitDim] = true

	leftNumBytes, err := w.recursePackIndex(writeBuffer, leafNodes, leftBlockFP, blocks,
		lastSplitValues, negativeDeltas, true, leavesOffset, numLeftLeafNodes)
	if err != nil {
		return 0, err
	}

	if numLeftLeafNodes != 1 {
		if err := writeBuffer.WriteUvarint(uint64(leftNumBytes)); err != nil {
			return 0, err
		}
	} else {
		//assert leftNumBytes == 0: "leftNumBytes=" + leftNumBytes;
	}

	bytes2 := slices.Clone(writeBuffer.Bytes())
	writeBuffer.Reset()
	// replace our placeholder:
	blocks.Set(idxSav, bytes2)

	negativeDeltas[splitDim] = false
	rightNumBytes, err := w.recursePackIndex(writeBuffer, leafNodes, leftBlockFP, blocks,
		lastSplitValues, negativeDeltas, false, rightOffset, numLeaves-numLeftLeafNodes)
	if err != nil {
		return 0, err
	}

	negativeDeltas[splitDim] = savNegativeDelta

	// restore lastSplitValues to what caller originally passed us:
	arraycopy(savSplitValue, 0, lastSplitValues, splitDim*config.bytesPerDim+prefix, suffix)

	return numBytes + len(bytes2) + leftNumBytes + rightNumBytes, nil
}

// Return an array that contains the min and max values for the [offset, offset+length] interval of the given BytesRefs.
func (w *Writer) computeMinMax(count int, packedValues func(int) []byte, offset, length int) ([][]byte, error) {
	minBuf := new(bytes.Buffer)
	maxBuf := new(bytes.Buffer)
	first := packedValues(0)

	minBuf.Write(first[offset : offset+length])
	maxBuf.Write(first[offset : offset+length])

	for i := 1; i < count; i++ {
		candidate := packedValues(i)[offset : offset+length]

		if bytes.Compare(minBuf.Bytes()[:length], candidate) > 0 {
			minBuf.Reset()
			minBuf.Write(candidate)
		} else if bytes.Compare(maxBuf.Bytes()[:length], candidate) < 0 {
			maxBuf.Reset()
			maxBuf.Write(candidate)
		}
	}
	return [][]byte{minBuf.Bytes(), maxBuf.Bytes()}, nil
}

func runLen(packedValues packedValuesFunc, start, end, byteOffset int) int {
	first := packedValues(start)
	b := first[byteOffset]
	for i := start + 1; i < end; i++ {
		ref := packedValues(i)
		b2 := ref[byteOffset]
		// assert Byte.toUnsignedInt(b2) >= Byte.toUnsignedInt(b);
		if b != b2 {
			return i - start
		}
	}
	return end - start
}

// Pick the next dimension to split.
// minPackedValue: the min values for all dimensions
// maxPackedValue: the max values for all dimensions
// parentSplits: how many times each dim has been split on the parent levels
// Returns: the dimension to split
func (w *Writer) split(minPackedValue, maxPackedValue []byte, parentSplits []int) (int, error) {
	// First look at whether there is a dimension that has split less than 2x less than
	// the dim that has most splits, and return it if there is such a dimension and it
	// does not only have equals values. This helps ensure all dimensions are indexed.
	maxNumSplits := 0
	for _, v := range parentSplits {
		maxNumSplits = max(maxNumSplits, v)
	}

	config := w.config
	numIndexDims := config.NumIndexDims()
	bytesPerDim := config.BytesPerDim()

	for dim := 0; dim < numIndexDims; dim++ {
		fromIndex := dim * config.bytesPerDim
		toIndex := fromIndex + config.bytesPerDim
		if parentSplits[dim] < maxNumSplits/2 &&
			compareUnsigned(minPackedValue, fromIndex, toIndex, maxPackedValue, fromIndex, toIndex) != 0 {
			return dim, nil
		}
	}

	// Find which dim has the largest span so we can split on it:
	splitDim := -1

	for dim := 0; dim < numIndexDims; dim++ {
		if err := Subtract(bytesPerDim, dim, maxPackedValue, minPackedValue, w.scratchDiff); err != nil {
			return 0, err
		}

		if splitDim == -1 ||
			compareUnsigned(w.scratchDiff, 0, config.bytesPerDim,
				w.scratch1, 0, config.bytesPerDim) > 0 {
			arraycopy(w.scratchDiff, 0, w.scratch1, 0, config.bytesPerDim)
			splitDim = dim
		}
	}

	return splitDim, nil
}

// Pull a partition back into heap once the point count is low enough while recursing.
func (w *Writer) switchToHeap(source PointWriter) (*HeapPointWriter, error) {
	count := source.Count()

	reader, err := source.GetReader(0, source.Count())
	if err != nil {
		return nil, err
	}
	writer := NewHeapPointWriter(w.config, count)
	for i := 0; i < count; i++ {
		if _, err := reader.Next(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		if err := writer.AppendPoint(reader.PointValue()); err != nil {
			return nil, err
		}
	}

	if err := source.Destroy(); err != nil {
		return nil, err
	}
	return writer, nil
}

func (w *Writer) computePackedValueBounds(slice *PathSlice, minPackedValue, maxPackedValue []byte) error {
	reader, err := slice.writer.GetReader(slice.start, slice.count)
	if err != nil {
		return err
	}
	if ok, err := reader.Next(); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	} else if !ok {
		return nil
	}

	value := reader.PointValue().PackedValue()

	config := w.config

	packedIndexBytesLength := config.PackedIndexBytesLength()
	bytesPerDim := config.BytesPerDim()
	numIndexDims := config.NumIndexDims()

	copy(minPackedValue, value[:packedIndexBytesLength])
	copy(maxPackedValue, value[:packedIndexBytesLength])

	for {
		next, err := reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		if !next {
			break
		}

		for dim := 0; dim < numIndexDims; dim++ {
			startOffset := dim * bytesPerDim
			endOffset := startOffset + bytesPerDim

			if bytes.Compare(value[startOffset:endOffset], minPackedValue[startOffset:endOffset]) < 0 {
				copy(minPackedValue[startOffset:endOffset], value[startOffset:endOffset])
			} else if bytes.Compare(value[startOffset:endOffset], maxPackedValue[startOffset:endOffset]) > 0 {
				copy(maxPackedValue[startOffset:endOffset], value[startOffset:endOffset])
			}
		}
	}
	return nil
}

func (w *Writer) computeCommonPrefixLength(heapPointWriter *HeapPointWriter, commonPrefix []byte, from, to int) {
	config := w.config

	for i := range w.commonPrefixLengths {
		w.commonPrefixLengths[i] = config.BytesPerDim()
	}
	value := heapPointWriter.GetPackedValueSlice(from)
	packedValue := value.PackedValue()

	for dim := 0; dim < config.NumDims(); dim++ {
		start := dim * config.BytesPerDim()
		end := start + config.BytesPerDim()
		copy(commonPrefix[start:], packedValue[start:end])
	}

	for i := from + 1; i < to; i++ {
		value = heapPointWriter.GetPackedValueSlice(i)
		packedValue = value.PackedValue()

		for dim := 0; dim < config.NumDims(); dim++ {
			if w.commonPrefixLengths[dim] != 0 {
				fromIndex := dim * config.BytesPerDim()
				toIndex := dim*config.BytesPerDim() + w.commonPrefixLengths[dim]

				j := Mismatch(commonPrefix[fromIndex:toIndex], packedValue[fromIndex:toIndex])
				if j != -1 {
					w.commonPrefixLengths[dim] = j
				}
			}
		}
	}
}
