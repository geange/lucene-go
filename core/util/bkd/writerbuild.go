package bkd

import (
	"bytes"
	"context"
	"math"
	"sort"

	"github.com/bits-and-blooms/bitset"
	"github.com/samber/lo"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

// Recursively reorders the provided reader and writes the bkd-tree on the fly; this method is used
// when we are writing a new segment directly from IndexWriter's indexing buffer (MutablePointsReader).
// 递归地重新排序所提供的读取器，并动态地写入bkd树；当我们直接从IndexWriter的索引缓冲区（MutablePointsReader）写入新段时，会使用此方法。
func (w *Writer) build(ctx context.Context, leavesOffset, numLeaves int, reader types.MutablePointValues, from, to int, out store.IndexOutput, minPackedValue, maxPackedValue []byte, parentSplits []int, splitPackedValues, splitDimensionValues []byte, leafBlockFPs []int64, spareDocIds []int) error {

	config := w.config

	numDims := config.NumDims()
	bytesPerDim := config.BytesPerDim()

	if numLeaves == 1 {
		return w.build1Leaf(ctx, leavesOffset, reader, from, to, out, leafBlockFPs, spareDocIds)
	}

	splitDim := 0
	if numDims != 1 {
		// for dimensions > 2 we recompute the bounds for the current inner node to help the algorithm choose best
		// split dimensions. Because it is an expensive operation, the frequency we recompute the bounds is given
		// by SPLITS_BEFORE_EXACT_BOUNDS.
		if numLeaves != len(leafBlockFPs) && config.NumIndexDims() > 2 &&
			lo.Sum(parentSplits)%SPLITS_BEFORE_EXACT_BOUNDS == 0 {
			w.computePackedValueBoundsV1(reader, from, to, minPackedValue, maxPackedValue, w.scratchBytesRef1)
		}

		var err error
		splitDim, err = w.split(minPackedValue, maxPackedValue, parentSplits)
		if err != nil {
			return err
		}
	}

	// How many leaves will be in the left tree:
	numLeftLeafNodes := w.getNumLeftLeafNodes(numLeaves)
	// How many points will be in the left tree:
	mid := from + numLeftLeafNodes*config.maxPointsInLeafNode

	startIndex := splitDim * bytesPerDim
	endIndex := startIndex + bytesPerDim
	commonPrefixLen := Mismatch(minPackedValue[startIndex:endIndex], maxPackedValue[startIndex:endIndex])
	if commonPrefixLen == -1 {
		commonPrefixLen = bytesPerDim
	}

	partition(config, w.maxDoc, splitDim, commonPrefixLen,
		reader, from, to, mid, w.scratchBytesRef1, w.scratchBytesRef2)

	rightOffset := leavesOffset + numLeftLeafNodes
	splitOffset := rightOffset - 1

	// set the split value
	address := splitOffset * bytesPerDim
	splitDimensionValues[splitOffset] = byte(splitDim)
	reader.GetValue(mid, w.scratchBytesRef1)

	start := splitDim * bytesPerDim
	end := start + bytesPerDim
	copy(splitPackedValues[address:], w.scratchBytesRef1.Bytes()[start:end])

	minSplitPackedValue := make([]byte, len(minPackedValue))
	maxSplitPackedValue := make([]byte, len(maxPackedValue))
	copy(minSplitPackedValue[start:end], w.scratchBytesRef1.Bytes()[start:end])
	copy(maxSplitPackedValue[start:end], w.scratchBytesRef1.Bytes()[start:end])

	// recurse
	parentSplits[splitDim]++
	if err := w.build(ctx, leavesOffset, numLeftLeafNodes, reader, from, mid, out, minPackedValue, maxSplitPackedValue, parentSplits, splitPackedValues, splitDimensionValues, leafBlockFPs, spareDocIds); err != nil {
		return err
	}

	if err := w.build(ctx, rightOffset, numLeaves-numLeftLeafNodes, reader, mid, to, out, minSplitPackedValue, maxPackedValue, parentSplits, splitPackedValues, splitDimensionValues, leafBlockFPs, spareDocIds); err != nil {
		return err
	}
	parentSplits[splitDim]--

	return nil
}

func (w *Writer) build1Leaf(ctx context.Context, leavesOffset int, reader types.MutablePointValues, from, to int, out store.IndexOutput, leafBlockFPs []int64, spareDocIds []int) error {

	config := w.config

	numDims := config.NumDims()
	bytesPerDim := config.BytesPerDim()

	count := to - from

	// Compute common prefixes
	for i := range w.commonPrefixLengths {
		w.commonPrefixLengths[i] = bytesPerDim
	}
	reader.GetValue(from, w.scratchBytesRef1)
	for i := from + 1; i < to; i++ {
		reader.GetValue(from, w.scratchBytesRef2)
		for dim := 0; dim < numDims; dim++ {
			start := dim * bytesPerDim
			end := start + bytesPerDim
			dimensionPrefixLength := w.commonPrefixLengths[dim]
			w.commonPrefixLengths[dim] = Mismatch(
				w.scratchBytesRef1.Bytes()[start:end],
				w.scratchBytesRef2.Bytes()[start:end],
			)

			if w.commonPrefixLengths[dim] == -1 {
				w.commonPrefixLengths[dim] = dimensionPrefixLength
			}
		}
	}

	// Find the dimension that has the least number of unique bytes at commonPrefixLengths[dim]
	usedBytes := make([]*bitset.BitSet, numDims)
	for dim := 0; dim < numDims; dim++ {
		if w.commonPrefixLengths[dim] < bytesPerDim {
			usedBytes[dim] = bitset.New(256)
		}
	}

	for i := from + 1; i < to; i++ {
		for dim := 0; dim < numDims; dim++ {
			if usedBytes[dim] != nil {
				b := reader.GetByteAt(i, dim*bytesPerDim+w.commonPrefixLengths[dim])
				usedBytes[dim].Set(uint(b))
			}
		}
	}

	sortedDim := 0
	sortedDimCardinality := math.MaxInt32
	for dim := 0; dim < numDims; dim++ {
		if usedBytes[dim] != nil {
			cardinality := int(usedBytes[dim].Count())
			if cardinality < sortedDimCardinality {
				sortedDim = dim
				sortedDimCardinality = cardinality
			}
		}
	}

	// sort by dim
	sort.Sort(NewIntroSorter(config, sortedDim, w.commonPrefixLengths,
		reader, from, to, w.scratchBytesRef1, w.scratchBytesRef2))

	comparator := w.scratchBytesRef1
	collector := w.scratchBytesRef2

	reader.GetValue(from, comparator)
	leafCardinality := 1

	for i := from + 1; i < to; i++ {
		reader.GetValue(i, collector)

		for dim := 0; dim < numDims; dim++ {
			start := dim*bytesPerDim + w.commonPrefixLengths[dim]
			end := dim*bytesPerDim + bytesPerDim

			cmp := bytes.Compare(collector.Bytes()[start:end], comparator.Bytes()[start:end])
			if cmp != -1 {
				leafCardinality++
				collector, comparator = comparator, collector
				break
			}
		}
	}

	// Save the block file pointer:
	leafBlockFPs[leavesOffset] = out.GetFilePointer()
	docIDs := spareDocIds

	for i := from; i < to; i++ {
		docIDs[i-from] = reader.GetDocID(i)
	}
	if err := w.writeLeafBlockDocs(ctx, w.scratchOut, docIDs[:count]); err != nil {
		return err
	}

	// Write the common prefixes:
	reader.GetValue(from, w.scratchBytesRef1)
	copy(w.scratch1, w.scratchBytesRef1.Bytes()[:config.packedBytesLength])
	if err := w.writeCommonPrefixes(ctx, w.scratchOut, w.commonPrefixLengths, w.scratch1); err != nil {
		return err
	}

	// Write the full values:
	packedValues := func(i int) []byte {
		reader.GetValue(from+i, w.scratchBytesRef1)
		return w.scratchBytesRef1.Bytes()
	}

	if err := w.writeLeafBlockPackedValues(w.scratchOut, w.commonPrefixLengths, count, sortedDim,
		packedValues, leafCardinality); err != nil {
		return err
	}

	if err := w.scratchOut.CopyTo(out); err != nil {
		return err
	}
	w.scratchOut.Reset()
	return nil
}

// The point writer contains the data that is going to be splitted using radix selection.
// This method is used when we are merging previously written segments, in the numDims > 1 case.
func (w *Writer) buildMerging(ctx context.Context, leavesOffset, numLeaves int, points *PathSlice, out store.IndexOutput, radixSelector *RadixSelector, minPackedValue, maxPackedValue []byte, parentSplits []int, splitPackedValues, splitDimensionValues []byte, leafBlockFPs []int64, spareDocIds []int) error {

	config := w.config
	//numIndexDims := config.NumIndexDims()
	//bytesPerDim := config.BytesPerDim()

	if numLeaves == 1 {
		return w.buildMerging1Leaf(ctx, leavesOffset, points, out, radixSelector, leafBlockFPs, spareDocIds)
	}

	// Inner node: partition/recurse

	var splitDim int
	var err error
	if config.numIndexDims == 1 {
		splitDim = 0
	} else {
		// for dimensions > 2 we recompute the bounds for the current inner node to help the algorithm choose best
		// split dimensions. Because it is an expensive operation, the frequency we recompute the bounds is given
		// by SPLITS_BEFORE_EXACT_BOUNDS.
		if numLeaves != len(leafBlockFPs) && config.numIndexDims > 2 && lo.Sum(parentSplits)%SPLITS_BEFORE_EXACT_BOUNDS == 0 {
			if err := w.computePackedValueBounds(points, minPackedValue, maxPackedValue); err != nil {
				return err
			}
		}
		splitDim, err = w.split(minPackedValue, maxPackedValue, parentSplits)
		if err != nil {
			return err
		}
	}

	//assert numLeaves <= leafBlockFPs.length : "numLeaves=" + numLeaves + " leafBlockFPs.length=" + leafBlockFPs.length;

	// How many leaves will be in the left tree:
	numLeftLeafNodes := w.getNumLeftLeafNodes(numLeaves)
	// How many points will be in the left tree:
	leftCount := numLeftLeafNodes * config.maxPointsInLeafNode

	slices := make([]*PathSlice, 2)

	commonPrefixLen := Mismatch(minPackedValue[splitDim*config.bytesPerDim:splitDim*config.bytesPerDim+config.bytesPerDim],
		maxPackedValue[splitDim*config.bytesPerDim:splitDim*config.bytesPerDim+config.bytesPerDim],
	)
	if commonPrefixLen == -1 {
		commonPrefixLen = config.bytesPerDim
	}

	splitValue, err := radixSelector.Select(points, slices, points.start, points.start+points.count, points.start+leftCount, splitDim, commonPrefixLen)
	if err != nil {
		return err
	}

	rightOffset := leavesOffset + numLeftLeafNodes
	splitValueOffset := rightOffset - 1

	splitDimensionValues[splitValueOffset] = byte(splitDim)
	address := splitValueOffset * config.bytesPerDim
	arraycopy(splitValue, 0, splitPackedValues, address, config.bytesPerDim)

	minSplitPackedValue := make([]byte, config.packedIndexBytesLength)
	arraycopy(minPackedValue, 0, minSplitPackedValue, 0, config.packedIndexBytesLength)

	maxSplitPackedValue := make([]byte, config.packedIndexBytesLength)
	arraycopy(maxPackedValue, 0, maxSplitPackedValue, 0, config.packedIndexBytesLength)

	arraycopy(splitValue, 0, minSplitPackedValue, splitDim*config.bytesPerDim, config.bytesPerDim)
	arraycopy(splitValue, 0, maxSplitPackedValue, splitDim*config.bytesPerDim, config.bytesPerDim)

	parentSplits[splitDim]++

	// Recurse on left tree:
	if err := w.buildMerging(ctx, leavesOffset, numLeftLeafNodes, slices[0], out, radixSelector, minPackedValue, maxSplitPackedValue, parentSplits, splitPackedValues, splitDimensionValues, leafBlockFPs, spareDocIds); err != nil {
		return err
	}

	// Recurse on right tree:
	if err := w.buildMerging(ctx, rightOffset, numLeaves-numLeftLeafNodes, slices[1], out, radixSelector, minSplitPackedValue, maxPackedValue, parentSplits, splitPackedValues, splitDimensionValues, leafBlockFPs, spareDocIds); err != nil {
		return err
	}

	parentSplits[splitDim]--
	return nil
}

func (w *Writer) buildMerging1Leaf(ctx context.Context, leavesOffset int, points *PathSlice, out store.IndexOutput, radixSelector *RadixSelector, leafBlockFPs []int64, spareDocIds []int) error {

	config := w.config

	// Leaf node: write block
	// We can write the block in any order so by default we write it sorted by the dimension that has the
	// least number of unique bytes at commonPrefixLengths[dim], which makes compression more efficient
	var heapSource *HeapPointWriter
	var err error

	if writer, ok := points.writer.(*HeapPointWriter); !ok {
		heapSource, err = w.switchToHeap(points.writer)
		if err != nil {
			return err
		}
	} else {
		heapSource = writer
	}

	from := points.Start()
	to := points.Start() + points.Count()

	//we store common prefix on scratch1
	w.computeCommonPrefixLength(heapSource, w.scratch1, from, to)

	sortedDim := 0
	sortedDimCardinality := math.MaxInt32
	usedBytes := make([]*bitset.BitSet, config.NumDims())
	for dim := 0; dim < config.NumDims(); dim++ {
		if w.commonPrefixLengths[dim] < config.BytesPerDim() {
			usedBytes[dim] = bitset.New(256)
		}
	}

	//Find the dimension to compress
	for dim := 0; dim < config.NumDims(); dim++ {
		prefix := w.commonPrefixLengths[dim]

		if prefix < config.BytesPerDim() {
			offset := dim * config.BytesPerDim()
			for i := from; i < to; i++ {
				value := heapSource.GetPackedValueSlice(i)
				packedValue := value.PackedValue()

				// 使用bitset来记录每个维度的基数（相似程度），相同的bucket越多，相似程度越高，占用的bitset的位数越少
				bucket := packedValue[offset+prefix]
				usedBytes[dim].Set(uint(bucket))
			}
			// 获取当前维度的基数值
			cardinality := int(usedBytes[dim].Count())
			if cardinality < sortedDimCardinality {
				// 使用基数最小的维度的作为排序的纬度
				sortedDim = dim
				sortedDimCardinality = cardinality
			}
		}
	}

	// sort the chosen dimension
	radixSelector.HeapRadixSort(heapSource, from, to, sortedDim, w.commonPrefixLengths[sortedDim])
	// compute cardinality
	leafCardinality := heapSource.ComputeCardinality(from, to, w.commonPrefixLengths)

	// Save the block file pointer:
	leafBlockFPs[leavesOffset] = out.GetFilePointer()

	// Write docIDs first, as their own chunk, so that at intersect time we can add all docIDs w/o
	// loading the values:
	count := to - from

	// Write doc IDs
	docIDs := spareDocIds
	for i := 0; i < count; i++ {
		docIDs[i] = heapSource.GetPackedValueSlice(from + i).DocID()
	}
	// 写入docID
	if err := w.writeLeafBlockDocs(ctx, out, docIDs[:count]); err != nil {
		return err
	}

	// TODO: minor opto: we don't really have to write the actual common prefixes,
	// because BKDReader on recursing can regenerate it for us
	// from the index, much like how terms dict does so from the FST:

	// Write the common prefixes:
	// 写入各个维度的公共前缀
	if err := w.writeCommonPrefixes(nil, out, w.commonPrefixLengths, w.scratch1); err != nil {
		return err
	}

	// Write the full values:
	packedValues := func() func(int) []byte {
		return func(i int) []byte {
			value := heapSource.GetPackedValueSlice(from + i)
			return value.PackedValue()
		}
	}

	// 写入各个维度的数据
	return w.writeLeafBlockPackedValues(out, w.commonPrefixLengths, count, sortedDim,
		packedValues(), leafCardinality)

}
