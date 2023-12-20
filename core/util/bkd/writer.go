package bkd

import (
	"bytes"
	"context"
	"errors"

	"slices"
	"sort"
	"sync/atomic"

	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/structure"
)

const (
	CODEC_NAME                     = "BKD"
	VERSION_START                  = 4
	VERSION_LEAF_STORES_BOUNDS     = 5
	VERSION_SELECTIVE_INDEXING     = 6
	VERSION_LOW_CARDINALITY_LEAVES = 7
	VERSION_META_FILE              = 9
	VERSION_CURRENT                = VERSION_META_FILE

	// SPLITS_BEFORE_EXACT_BOUNDS Number of splits before we compute the exact bounding box of an inner node.
	SPLITS_BEFORE_EXACT_BOUNDS = 4

	// DEFAULT_MAX_MB_SORT_IN_HEAP Default maximum heap to use, before spilling to (slower) disk
	DEFAULT_MAX_MB_SORT_IN_HEAP = 16.0
)

// TODO
//   - allow variable length byte[] (across docs and dims), but this is quite a bit more hairy
//   - we could also index "auto-prefix terms" here, and use better compression, and maybe only use for the "fully contained" case so we'd
//     only index docIDs
//   - the index could be efficiently encoded as an FST, so we don't have wasteful
//     (monotonic) long[] leafBlockFPs; or we could use MonotonicLongValues ... but then
//     the index is already plenty small: 60M OSM points --> 1.1 MB with 128 points
//     per leaf, and you can reduce that by putting more points per leaf
//   - we could use threads while building; the higher nodes are very parallelizable

// Writer Recursively builds a block KD-tree to assign all incoming points in N-dim space to smaller
// and smaller N-dim rectangles (cells) until the number of points in a given rectangle
// is <= config.maxPointsInLeafNode. The tree is partially balanced, which means the leaf nodes will
// have the requested config.maxPointsInLeafNode except one that might have less. Leaf nodes may
// straddle the two bottom levels of the binary tree. Values that fall exactly on a cell boundary
// may be in either cell.
// The number of dimensions can be 1 to 8, but every byte[] value is fixed length.
// This consumes heap during writing: it allocates a Long[numLeaves], a byte[numLeaves*(1+config.bytesPerDim)]
// and then uses up to the specified maxMBSortInHeap heap space for writing.
// NOTE: This can write at most Integer.MAX_VALUE * config.maxPointsInLeafNode / config.bytesPerDim total points.
// lucene.experimental
type Writer struct {
	config              *Config
	tempDir             *store.TrackingDirectoryWrapper
	tempFileNamePrefix  string
	maxMBSortInHeap     float64
	scratchDiff         []byte
	scratch1            []byte
	scratch2            []byte
	scratchBytesRef1    *bytes.Buffer
	scratchBytesRef2    *bytes.Buffer
	commonPrefixLengths []int
	docsSeen            *bitset.BitSet
	pointWriter         PointWriter
	finished            *atomic.Bool
	tempInput           store.IndexOutput
	maxPointsSortInHeap int
	minPackedValue      []byte
	maxPackedValue      []byte
	pointCount          int
	totalPointCount     int
	maxDoc              int

	// Reused when writing leaf blocks
	scratchOut *store.ByteBuffersDataOutput
}

func NewWriter(maxDoc int, tempDir store.Directory, tempFileNamePrefix string,
	config *Config, maxMBSortInHeap float64, totalPointCount int) (*Writer, error) {

	writer := &Writer{
		config:              config,
		tempDir:             store.NewTrackingDirectoryWrapper(tempDir),
		tempFileNamePrefix:  tempFileNamePrefix,
		maxMBSortInHeap:     maxMBSortInHeap,
		scratchDiff:         make([]byte, config.BytesPerDim()),
		scratch1:            make([]byte, config.PackedBytesLength()),
		scratch2:            make([]byte, config.PackedBytesLength()),
		scratchBytesRef1:    new(bytes.Buffer),
		scratchBytesRef2:    new(bytes.Buffer),
		commonPrefixLengths: make([]int, config.NumDims()),
		docsSeen:            bitset.New(uint(maxDoc)),
		pointWriter:         nil,
		finished:            &atomic.Bool{},
		tempInput:           nil,
		maxPointsSortInHeap: int((maxMBSortInHeap)*1024*1024) / (config.BytesPerDoc()),
		minPackedValue:      make([]byte, config.PackedIndexBytesLength()),
		maxPackedValue:      make([]byte, config.PackedIndexBytesLength()),
		pointCount:          0,
		totalPointCount:     totalPointCount,
		maxDoc:              maxDoc,
		scratchOut:          nil,
	}

	return writer, nil
}

func (w *Writer) init() error {
	if w.pointWriter != nil {
		return errors.New("pointWriter not nil")
	}

	config := w.config

	if w.totalPointCount > w.maxPointsSortInHeap {
		writer := NewOfflinePointWriter(config, w.tempDir, w.tempFileNamePrefix, "spill", 0)
		w.pointWriter = writer
		w.tempInput = writer.out
	} else {
		w.pointWriter = NewHeapPointWriter(config, w.totalPointCount)
	}
	return nil
}

func (w *Writer) Add(packedValue []byte, docID int) error {
	config := w.config

	if len(packedValue) != config.PackedIndexBytesLength() {
		return errors.New("packedValue length not equal")
	}

	if w.pointCount >= w.totalPointCount {
		return errors.New("too many points")
	}

	if w.pointCount == 0 {
		if err := w.init(); err != nil {
			return err
		}
		//copy(w.minPackedValue, packedValue[:config.packedIndexBytesLength])
		//copy(w.maxPackedValue, packedValue[:config.packedIndexBytesLength])
		arraycopy(packedValue, 0, w.minPackedValue, 0, config.packedIndexBytesLength)
		arraycopy(packedValue, 0, w.maxPackedValue, 0, config.packedIndexBytesLength)
	} else {
		for dim := 0; dim < config.NumIndexDims(); dim++ {
			//start := dim * config.BytesPerDim()
			//end := start + config.BytesPerDim()
			//packed := packedValue[start:end]

			//if bytes.Compare(packed, w.minPackedValue[start:end]) < 0 {
			//	copy(w.minPackedValue[start:end], packed)
			//	continue
			//}
			//
			//if bytes.Compare(packed, w.maxPackedValue[start:end]) > 0 {
			//	copy(w.maxPackedValue[start:end], packed)
			//	continue
			//}

			offset := dim * config.bytesPerDim
			if compareUnsigned(packedValue, offset, offset+config.bytesPerDim, w.minPackedValue, offset, offset+config.bytesPerDim) < 0 {
				arraycopy(packedValue, offset, w.minPackedValue, offset, config.bytesPerDim)
			} else if compareUnsigned(packedValue, offset, offset+config.bytesPerDim, w.maxPackedValue, offset, offset+config.bytesPerDim) > 0 {
				arraycopy(packedValue, offset, w.maxPackedValue, offset, config.bytesPerDim)
			}
		}
	}

	if err := w.pointWriter.Append(nil, packedValue, docID); err != nil {
		return err
	}
	w.pointCount++
	w.docsSeen.Set(uint(docID))
	return nil
}

type MergeQueue struct {
	structure.PriorityQueue[MergeReader]

	bytesPerDim int
}

// LeafNodes flat representation of a kd-tree
type LeafNodes interface {
	// NumLeaves number of leaf nodes
	// 叶子节点的数量
	NumLeaves() int

	// GetLeafLP pointer to the leaf node previously written.
	// Leaves are order from left to right,
	// so leaf at index 0 is the leftmost leaf and the the leaf at numleaves() -1 is the rightmost leaf
	// 指向先前写入的叶节点的指针。叶子是从左到右的顺序，所以索引0处的叶子是最左边的叶子，NumLeaves()-1 处的叶子则是最右边的叶子
	GetLeafLP(index int) int64

	// GetSplitValue split value between two leaves.
	// The split value at position n corresponds to the leaves at (n -1) and n.
	// 两片叶子之间的分割值。位置n处的分割值对应于（n-1）和n处的叶子。
	GetSplitValue(index int) []byte

	// GetSplitDimension split dimension between two leaves.
	// The split dimension at position n corresponds to the leaves at (n -1) and n.
	// 两片叶子之间的分割的维度。位置n处的分割尺寸对应于（n-1）和n处的叶子。
	GetSplitDimension(index int) int
}

type Runnable func(ctx context.Context) error

var (
	emptyRunnable = func(_ context.Context) error { return nil }
)

// WriteField Write a field from a MutablePointValues. This way of writing points is faster than regular writes with add since there is opportunity for reordering points before writing them to disk. This method does not use transient disk in order to reorder points.
func (w *Writer) WriteField(ctx context.Context, metaOut, indexOut, dataOut store.IndexOutput, fieldName string, reader types.MutablePointValues) (Runnable, error) {
	if w.config.NumDims() == 1 {
		return w.writeField1Dim(metaOut, indexOut, dataOut, fieldName, reader)
	} else {
		return w.writeFieldNDims(ctx, metaOut, indexOut, dataOut, fieldName, reader)
	}
}

func (w *Writer) computePackedValueBoundsV1(values types.MutablePointValues, from, to int,
	minPackedValue, maxPackedValue []byte, scratch *bytes.Buffer) {

	if from == to {
		return
	}

	config := w.config

	values.GetValue(from, scratch)
	copy(minPackedValue, scratch.Bytes()[:config.PackedIndexBytesLength()])
	copy(maxPackedValue, scratch.Bytes()[:config.PackedIndexBytesLength()])

	for i := from + 1; i < to; i++ {
		values.GetValue(from, scratch)

		for dim := 0; dim < config.NumIndexDims(); dim++ {
			start := dim * config.BytesPerDim()
			end := start + config.BytesPerDim()

			value := scratch.Bytes()[start:end]

			if bytes.Compare(value, minPackedValue[start:end]) < 0 {
				copy(minPackedValue[start:end], value)
				continue
			}

			if bytes.Compare(value, maxPackedValue[start:end]) > 0 {
				copy(maxPackedValue[start:end], value)
				continue
			}
		}
	}
}

// In the 2+D case, we recursively pick the split dimension, compute the
// median value and partition other values around it.
func (w *Writer) writeFieldNDims(ctx context.Context, metaOut, indexOut, dataOut store.IndexOutput, fieldName string, values types.MutablePointValues) (Runnable, error) {

	config := w.config

	if w.pointCount != 0 {
		return nil, errors.New("cannot mix add and writeField")
	}

	// Catch user silliness:
	if w.finished.Load() {
		return nil, errors.New("already finished")
	}

	// Mark that we already finished:
	w.finished.Store(true)

	w.pointCount = values.Size()

	numLeaves := (w.pointCount + config.MaxPointsInLeafNode() - 1) / config.MaxPointsInLeafNode()

	numSplits := numLeaves - 1

	splitPackedValues := make([]byte, numSplits*config.BytesPerDim())
	splitDimensionValues := make([]byte, numSplits)
	leafBlockFPs := make([]int64, numLeaves)

	// compute the min/max for this slice
	w.computePackedValueBoundsV1(values, 0, w.pointCount,
		w.minPackedValue, w.maxPackedValue, w.scratchBytesRef1)

	for i := 0; i < w.pointCount; i++ {
		w.docsSeen.Set(uint(values.GetDocID(i)))
	}

	dataStartFP := dataOut.GetFilePointer()
	parentSplits := make([]int, config.NumIndexDims())

	minPackedValueCopy := make([]byte, len(w.minPackedValue))
	copy(minPackedValueCopy, w.minPackedValue)

	maxPackedValueCopy := make([]byte, len(w.maxPackedValue))
	copy(maxPackedValueCopy, w.maxPackedValue)

	if err := w.build(ctx, 0, numLeaves, values, 0, w.pointCount, dataOut, minPackedValueCopy, maxPackedValueCopy, parentSplits, splitPackedValues, splitDimensionValues, leafBlockFPs, make([]int, config.MaxPointsInLeafNode())); err != nil {
		return nil, err
	}

	leafNodes := &leafNodesWriteFieldNDims{
		leafBlockFPs:         leafBlockFPs,
		p:                    w,
		splitDimensionValues: splitDimensionValues,
	}

	return func(ctx context.Context) error {
		return w.writeIndex(ctx, metaOut, indexOut, config.MaxPointsInLeafNode(), leafNodes, dataStartFP)
	}, nil
}

var _ LeafNodes = &leafNodesWriteFieldNDims{}

type leafNodesWriteFieldNDims struct {
	leafBlockFPs         []int64
	p                    *Writer
	splitDimensionValues []byte
}

func (l *leafNodesWriteFieldNDims) NumLeaves() int {
	return len(l.leafBlockFPs)
}

func (l *leafNodesWriteFieldNDims) GetLeafLP(index int) int64 {
	return l.leafBlockFPs[index]
}

func (l *leafNodesWriteFieldNDims) GetSplitValue(index int) []byte {
	offset := index * l.p.config.bytesPerDim
	return l.p.scratchBytesRef1.Bytes()[offset:]
}

func (l *leafNodesWriteFieldNDims) GetSplitDimension(index int) int {
	return int(l.splitDimensionValues[index])
}

// In the 1D case, we can simply sort points in ascending order and use the
// same writing logic as we use at merge time.
func (w *Writer) writeField1Dim(metaOut, indexOut, dataOut store.IndexOutput,
	fieldName string, reader types.MutablePointValues) (Runnable, error) {

	sorter := NewMutablePointValuesSorter(w.config, w.maxDoc, reader, 0, reader.Size())
	sort.Sort(sorter)

	oneDimWriter, err := w.NewOneDimensionBKDWriter(metaOut, indexOut, dataOut)
	if err != nil {
		return nil, err
	}

	err = reader.Intersect(nil, &writeField1DimVisitor{oneDimWriter: oneDimWriter})
	if err != nil {
		return nil, err
	}

	return oneDimWriter.Finish()
}

var _ types.IntersectVisitor = &writeField1DimVisitor{}

type writeField1DimVisitor struct {
	oneDimWriter *OneDimensionBKDWriter
}

func (w *writeField1DimVisitor) Visit(docID int) error {
	return errors.New("IllegalStateException")
}

func (w *writeField1DimVisitor) VisitLeaf(docID int, packedValue []byte) error {
	return w.oneDimWriter.Add(packedValue, docID)
}

func (w *writeField1DimVisitor) Compare(minPackedValue, maxPackedValue []byte) types.Relation {
	return types.CELL_CROSSES_QUERY
}

func (w *writeField1DimVisitor) Grow(count int) {
	return
}

// More efficient bulk-add for incoming BKDReaders. This does a merge sort of the already sorted values and currently only works when numDims==1. This returns -1 if all documents containing dimensional values were deleted.
func (w *Writer) merge(metaOut, indexOut, dataOut store.IndexOutput,
	docMaps []types.DocMap, readers []*Reader) (Runnable, error) {

	panic("")
}

func (w *Writer) getNumLeftLeafNodes(numLeaves int) int {
	// assert numLeaves > 1 : "getNumLeftLeaveNodes() called with " + numLeaves;
	// return the level that can be filled with this number of leaves
	lastFullLevel := 31 - numberOfLeadingZeros(int32(numLeaves))
	// how many leaf nodes are in the full level
	leavesFullLevel := 1 << lastFullLevel
	// half of the leaf nodes from the full level goes to the left
	numLeftLeafNodes := leavesFullLevel / 2
	// leaf nodes that do not fit in the full level
	unbalancedLeafNodes := numLeaves - leavesFullLevel
	// distribute unbalanced leaf nodes
	numLeftLeafNodes += min(int(unbalancedLeafNodes), int(numLeftLeafNodes))
	// we should always place unbalanced leaf nodes on the left
	// assert numLeftLeafNodes >= numLeaves - numLeftLeafNodes && numLeftLeafNodes <= 2L * (numLeaves - numLeftLeafNodes);
	return numLeftLeafNodes
}

// Finish Writes the BKD tree to the provided IndexOutputs and returns a Runnable that writes the index of the tree
// if at least one point has been added, or null otherwise.
func (w *Writer) Finish(ctx context.Context, metaOut, indexOut, dataOut store.IndexOutput) (Runnable, error) {
	config := w.config
	bytesPerDim := config.BytesPerDim()

	if w.finished.Load() == true {
		return nil, errors.New("already finished")
	}

	if w.pointCount == 0 {
		return emptyRunnable, nil
	}

	w.finished.Store(true)

	err := w.pointWriter.Close()
	if err != nil {
		return nil, err
	}

	pathSlice := NewPathSlice(w.pointWriter, 0, w.pointCount)
	w.tempInput = nil
	w.pointWriter = nil

	numLeaves := (w.pointCount + config.MaxPointsInLeafNode() - 1) / config.MaxPointsInLeafNode()
	numSplits := numLeaves - 1

	// checkMaxLeafNodeCount(numLeaves

	// NOTE: we could save the 1+ here, to use a bit less heap at search time, but then we'd need a somewhat costly check at each
	// step of the recursion to recompute the split dim:

	// Indexed by nodeID, but first (root) nodeID is 1.  We do 1+ because the lead byte at each recursion says which dim we split on.
	splitPackedValues := make([]byte, numSplits*bytesPerDim)
	splitDimensionValues := make([]byte, numSplits)

	// +1 because leaf count is power of 2 (e.g. 8), and innerNodeCount is power of 2 minus 1 (e.g. 7)
	leafBlockFPs := make([]int64, numLeaves)

	//We re-use the selector so we do not need to create an object every time.
	selector := NewRadixSelector(config, w.maxPointsSortInHeap, w.tempDir, w.tempFileNamePrefix)

	dataStartFP := dataOut.GetFilePointer()

	parentSplits := make([]int, config.NumIndexDims())

	err = w.buildMerging(ctx, 0, numLeaves, pathSlice, dataOut, selector, slices.Clone(w.minPackedValue), slices.Clone(w.maxPackedValue), parentSplits, splitPackedValues, splitDimensionValues, leafBlockFPs, make([]int, config.MaxPointsInLeafNode()))
	if err != nil {
		return nil, err
	}

	leafNodes := &leafNodesOnFinish{
		config:               config,
		leafBlockFPs:         leafBlockFPs,
		splitPackedValues:    splitPackedValues,
		splitDimensionValues: splitDimensionValues,
	}

	return func(ctx context.Context) error {
		return w.writeIndex(ctx, metaOut, indexOut, config.MaxPointsInLeafNode(), leafNodes, dataStartFP)
	}, nil
}

var _ LeafNodes = &leafNodesOnFinish{}

type leafNodesOnFinish struct {
	config               *Config
	leafBlockFPs         []int64
	splitPackedValues    []byte
	splitDimensionValues []byte
}

func (r *leafNodesOnFinish) NumLeaves() int {
	return len(r.leafBlockFPs)
}

func (r *leafNodesOnFinish) GetLeafLP(index int) int64 {
	return r.leafBlockFPs[index]
}

func (r *leafNodesOnFinish) GetSplitValue(index int) []byte {
	fromIndex := index * r.config.BytesPerDim()
	toIndex := fromIndex + r.config.BytesPerDim()
	return r.splitPackedValues[fromIndex:toIndex]
}

func (r *leafNodesOnFinish) GetSplitDimension(index int) int {
	return int(r.splitDimensionValues[index])
}

func (w *Writer) Close() error {
	w.finished.Store(true)
	if w.tempInput != nil {
		if err := w.tempInput.Close(); err != nil {
			return err
		}
		if err := w.tempDir.DeleteFile(w.tempInput.GetName()); err != nil {
			return err
		}
		w.tempInput = nil
	}
	return nil
}
