package bkd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"slices"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

var _ types.PointValues = &Reader{}

// Reader
// Handles intersection of an multi-dimensional shape in byte[] space with a block KD-tree previously
// written with BKDWriter.
// lucene.experimental
type Reader struct {
	// Packed array of byte[] holding all split values in the full binary tree:
	leafNodeOffset int
	config         *Config
	numLeaves      int
	in             store.IndexInput
	minPackedValue []byte
	maxPackedValue []byte
	pointCount     int
	docCount       int
	version        int
	minLeafBlockFP int64

	packedIndex store.IndexInput
}

// NewReader
// Caller must pre-seek the provided IndexInput to the index location that BKDWriter.finish returned.
// BKD tree is always stored off-heap.
func NewReader(ctx context.Context, metaIn, indexIn, dataIn store.IndexInput) (*Reader, error) {
	version, err := utils.CheckHeader(ctx, metaIn, CODEC_NAME, VERSION_START, VERSION_CURRENT)
	if err != nil {
		return nil, err
	}

	numDims, err := metaIn.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	var numIndexDims int
	if version >= VERSION_SELECTIVE_INDEXING {
		n, err := metaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		numIndexDims = int(n)
	} else {
		numIndexDims = int(numDims)
	}

	maxPointsInLeafNode, err := metaIn.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}

	bytesPerDim, err := metaIn.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}

	config, err := NewConfig(int(numDims), numIndexDims, int(bytesPerDim), int(maxPointsInLeafNode))
	if err != nil {
		return nil, err
	}

	// Read index:
	numLeaves, err := metaIn.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}

	packedIndexBytesLength := config.PackedIndexBytesLength()
	minPackedValue := make([]byte, packedIndexBytesLength)
	maxPackedValue := make([]byte, packedIndexBytesLength)

	_, err = metaIn.Read(minPackedValue)
	if err != nil {
		return nil, err
	}
	_, err = metaIn.Read(maxPackedValue)
	if err != nil {
		return nil, err
	}

	// 校验维度
	for dim := 0; dim < config.NumIndexDims(); dim++ {
		fromIndex := dim * config.BytesPerDim()
		toIndex := fromIndex + config.BytesPerDim()

		if bytes.Compare(minPackedValue[fromIndex:toIndex], maxPackedValue[fromIndex:toIndex]) > 0 {
			err := fmt.Errorf("minPackedValue is bigger than maxPackedValue")
			return nil, err
		}
	}

	pointCount, err := metaIn.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	docCount, err := metaIn.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}

	numIndexBytes, err := metaIn.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	var indexStartPointer int64
	var minLeafBlockFP int64
	if version >= VERSION_META_FILE {
		minFP, err := metaIn.ReadUint64(ctx)
		if err != nil {
			return nil, err
		}
		minLeafBlockFP = int64(minFP)
		startPointer, err := metaIn.ReadUint64(ctx)
		if err != nil {
			return nil, err
		}
		indexStartPointer = int64(startPointer)
	} else {
		indexStartPointer = indexIn.GetFilePointer()
		minFP, err := indexIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		minLeafBlockFP = int64(minFP)
		if _, err = indexIn.Seek(indexStartPointer, io.SeekStart); err != nil {
			return nil, err
		}
	}

	packedIndex, err := indexIn.Slice("packedIndex", indexStartPointer, int64(numIndexBytes))
	if err != nil {
		return nil, err
	}

	reader := &Reader{
		leafNodeOffset: int(numLeaves),
		config:         config,
		numLeaves:      int(numLeaves),
		in:             dataIn,
		minPackedValue: minPackedValue,
		maxPackedValue: maxPackedValue,
		pointCount:     int(pointCount),
		docCount:       int(docCount),
		version:        version,
		minLeafBlockFP: minLeafBlockFP,
		packedIndex:    packedIndex,
	}

	return reader, nil
}

func (r *Reader) Intersect(ctx context.Context, visitor types.IntersectVisitor) error {
	state, err := r.GetIntersectState(ctx, visitor)
	if err != nil {
		return err
	}

	return r.intersect(nil, state, r.minPackedValue, r.maxPackedValue)
}

func (r *Reader) EstimatePointCount(ctx context.Context, visitor types.IntersectVisitor) (int, error) {
	state, err := r.GetIntersectState(ctx, visitor)
	if err != nil {
		return -1, nil
	}

	return r.estimatePointCount(ctx, state, r.minPackedValue, r.maxPackedValue)
}

func (r *Reader) GetIntersectState(ctx context.Context, visitor types.IntersectVisitor) (*IntersectState, error) {
	index, err := r.newIndexTree(ctx)
	if err != nil {
		return nil, err
	}
	return NewIntersectState(r.in.Clone().(store.IndexInput), r.config, visitor, index), nil
}

// VisitLeafBlockValues Visits all docIDs and packed values in a single leaf block
func (r *Reader) VisitLeafBlockValues(ctx context.Context, index *IndexTree, state *IntersectState) error {
	// Leaf node; scan and filter all points in this block:
	count, err := r.readDocIDs(ctx, state.in, index.GetLeafBlockFP(), state.scratchIterator)
	if err != nil {
		return err
	}

	// Again, this time reading values and checking with the visitor
	return r.visitDocValues(ctx, state.commonPrefixLengths, state.scratchDataPackedValue, state.scratchMinIndexPackedValue, state.scratchMaxIndexPackedValue, state.in, state.scratchIterator, count, state.visitor)
}

func (r *Reader) visitDocIDs(ctx context.Context, in store.IndexInput, blockFP int64, visitor types.IntersectVisitor) error {
	// Leaf node
	if _, err := in.Seek(blockFP, io.SeekStart); err != nil {
		return err
	}

	// How many points are stored in this leaf cell:
	count, err := in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	// No need to call grow(), it has been called up-front

	return ReadIntsVisitor(ctx, in, int(count), visitor)
}

func (r *Reader) readDocIDs(ctx context.Context, in store.IndexInput, blockFP int64, iterator *readerDocIDSetIterator) (int, error) {
	if _, err := in.Seek(blockFP, io.SeekStart); err != nil {
		return 0, err
	}

	// How many points are stored in this leaf cell:
	count, err := in.ReadUvarint(ctx)
	if err != nil {
		return 0, err
	}

	if err = ReadInts(nil, in, int(count), iterator.docIDs); err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *Reader) visitDocValues(ctx context.Context, commonPrefixLengths []int, scratchDataPackedValue, scratchMinIndexPackedValue, scratchMaxIndexPackedValue []byte, in store.IndexInput, scratchIterator *readerDocIDSetIterator, count int, visitor types.IntersectVisitor) error {

	if r.version >= VERSION_LOW_CARDINALITY_LEAVES {
		return r.visitDocValuesWithCardinality(ctx, commonPrefixLengths, scratchDataPackedValue, scratchMinIndexPackedValue, scratchMaxIndexPackedValue, in, scratchIterator, count, visitor)
	} else {
		return r.visitDocValuesNoCardinality(ctx, commonPrefixLengths, scratchDataPackedValue, scratchMinIndexPackedValue, scratchMaxIndexPackedValue, in, scratchIterator, count, visitor)
	}
}

func (r *Reader) visitDocValuesNoCardinality(ctx context.Context, commonPrefixLengths []int, scratchDataPackedValue, scratchMinIndexPackedValue, scratchMaxIndexPackedValue []byte, in store.IndexInput, scratchIterator *readerDocIDSetIterator, count int, visitor types.IntersectVisitor) error {

	config := r.config

	if config.numIndexDims != 1 && r.version >= VERSION_LEAF_STORES_BOUNDS {
		minPackedValue := scratchMinIndexPackedValue
		arraycopy(scratchDataPackedValue, 0, minPackedValue, 0, config.packedIndexBytesLength)
		maxPackedValue := scratchMaxIndexPackedValue
		// Copy common prefixes before reading adjusted box
		arraycopy(minPackedValue, 0, maxPackedValue, 0, config.packedIndexBytesLength)
		if err := r.readMinMax(commonPrefixLengths, minPackedValue, maxPackedValue, in); err != nil {
			return err
		}

		// The index gives us range of values for each dimension, but the actual range of values
		// might be much more narrow than what the index told us, so we double check the relation
		// here, which is cheap yet might help figure out that the block either entirely matches
		// or does not match at all. This is especially more likely in the case that there are
		// multiple dimensions that have correlation, ie. splitting on one dimension also
		// significantly changes the range of values in another dimension.
		relation := visitor.Compare(minPackedValue, maxPackedValue)
		if relation == types.CELL_OUTSIDE_QUERY {
			return nil
		}
		visitor.Grow(count)

		if relation == types.CELL_INSIDE_QUERY {
			for i := 0; i < count; i++ {
				if err := visitor.Visit(ctx, scratchIterator.docIDs[i]); err != nil {
					return err
				}
			}
			return nil
		}
	} else {
		visitor.Grow(count)
	}

	compressedDim, err := r.readCompressedDim(in)
	if err != nil {
		return err
	}

	if compressedDim == -1 {
		if err := r.visitUniqueRawDocValues(scratchDataPackedValue, scratchIterator, count, visitor); err != nil {
			return err
		}
	} else {
		if err := r.visitCompressedDocValues(ctx, commonPrefixLengths, scratchDataPackedValue, in, scratchIterator, count, visitor, compressedDim); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reader) visitDocValuesWithCardinality(ctx context.Context, commonPrefixLengths []int, scratchDataPackedValue, scratchMinIndexPackedValue, scratchMaxIndexPackedValue []byte, in store.IndexInput, scratchIterator *readerDocIDSetIterator, count int, visitor types.IntersectVisitor) error {

	if err := r.readCommonPrefixes(ctx, commonPrefixLengths, scratchDataPackedValue, in); err != nil {
		return err
	}

	compressedDim, err := r.readCompressedDim(in)
	if err != nil {
		return err
	}
	if compressedDim == -1 {
		// all values are the same
		visitor.Grow(count)
		if err := r.visitUniqueRawDocValues(scratchDataPackedValue, scratchIterator, count, visitor); err != nil {
			return err
		}
	} else {
		config := r.config

		if config.numIndexDims != 1 {
			minPackedValue := scratchMinIndexPackedValue
			arraycopy(scratchDataPackedValue, 0, minPackedValue, 0, config.packedIndexBytesLength)
			maxPackedValue := scratchMaxIndexPackedValue
			// Copy common prefixes before reading adjusted box
			arraycopy(minPackedValue, 0, maxPackedValue, 0, config.packedIndexBytesLength)

			if err := r.readMinMax(commonPrefixLengths, minPackedValue, maxPackedValue, in); err != nil {
				return err
			}

			// The index gives us range of values for each dimension, but the actual range of values
			// might be much more narrow than what the index told us, so we double check the relation
			// here, which is cheap yet might help figure out that the block either entirely matches
			// or does not match at all. This is especially more likely in the case that there are
			// multiple dimensions that have correlation, ie. splitting on one dimension also
			// significantly changes the range of values in another dimension.
			relation := visitor.Compare(minPackedValue, maxPackedValue)
			if relation == types.CELL_OUTSIDE_QUERY {
				return nil
			}
			visitor.Grow(count)

			if relation == types.CELL_INSIDE_QUERY {
				for i := 0; i < count; i++ {
					if err := visitor.Visit(ctx, scratchIterator.docIDs[i]); err != nil {
						return err
					}
				}
				return nil
			}
		} else {
			visitor.Grow(count)
		}
		if compressedDim == -2 {
			// low cardinality values
			if err := r.visitSparseRawDocValues(ctx, commonPrefixLengths, scratchDataPackedValue, in, scratchIterator, count, visitor); err != nil {
				return err
			}
		} else {
			// high cardinality
			if err := r.visitCompressedDocValues(ctx, commonPrefixLengths, scratchDataPackedValue, in, scratchIterator, count, visitor, compressedDim); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Reader) readMinMax(commonPrefixLengths []int, minPackedValue, maxPackedValue []byte, in store.IndexInput) error {
	config := r.config
	for dim := 0; dim < config.numIndexDims; dim++ {
		prefix := commonPrefixLengths[dim]

		from := dim*config.bytesPerDim + prefix
		to := from + config.bytesPerDim - prefix

		if _, err := in.Read(minPackedValue[from:to]); err != nil {
			return err
		}
		if _, err := in.Read(maxPackedValue[from:to]); err != nil {
			return err
		}
	}
	return nil
}

// read cardinality and point
func (r *Reader) visitSparseRawDocValues(ctx context.Context, commonPrefixLengths []int, scratchPackedValue []byte, in store.IndexInput, scratchIterator *readerDocIDSetIterator, count int, visitor types.IntersectVisitor) error {

	config := r.config

	var i int
	for i = 0; i < count; {
		length, err := in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		for dim := 0; dim < config.numDims; dim++ {
			prefix := commonPrefixLengths[dim]

			from := dim*config.bytesPerDim + prefix
			to := from + config.bytesPerDim - prefix

			if _, err := in.Read(scratchPackedValue[from:to]); err != nil {
				return err
			}
		}
		scratchIterator.reset(i, int(length))
		if err := types.Visit(ctx, visitor, scratchIterator, scratchPackedValue); err != nil {
			return err
		}
		i += int(length)
	}
	if i != count {
		return fmt.Errorf("sub blocks do not add up to the expected count")
	}
	return nil
}

// point is under commonPrefix
func (r *Reader) visitUniqueRawDocValues(scratchPackedValue []byte,
	scratchIterator *readerDocIDSetIterator, count int, visitor types.IntersectVisitor) error {
	scratchIterator.reset(0, count)
	return types.Visit(nil, visitor, scratchIterator, scratchPackedValue)
}

func (r *Reader) visitCompressedDocValues(ctx context.Context, commonPrefixLengths []int, scratchPackedValue []byte, in store.IndexInput, scratchIterator *readerDocIDSetIterator, count int, visitor types.IntersectVisitor, compressedDim int) error {

	config := r.config

	// the byte at `compressedByteOffset` is compressed using run-length compression,
	// other suffix bytes are stored verbatim
	compressedByteOffset := compressedDim*config.bytesPerDim + commonPrefixLengths[compressedDim]
	commonPrefixLengths[compressedDim]++

	var i int
	for i = 0; i < count; {
		b1, err := in.ReadByte()
		if err != nil {
			return err
		}
		scratchPackedValue[compressedByteOffset] = b1

		b2, err := in.ReadByte()
		if err != nil {
			return err
		}

		runLength := int(b2)
		for j := 0; j < runLength; j++ {
			for dim := 0; dim < config.numDims; dim++ {
				prefix := commonPrefixLengths[dim]

				from := dim*config.bytesPerDim + prefix
				to := from + config.bytesPerDim - prefix

				if _, err := in.Read(scratchPackedValue[from:to]); err != nil {
					return err
				}
			}

			if err := visitor.VisitLeaf(ctx, scratchIterator.docIDs[i+j], scratchPackedValue); err != nil {
				return err
			}

		}
		i += runLength
	}
	if i != count {
		return fmt.Errorf("sub blocks do not add up to the expected count")
	}

	return nil
}

func (r *Reader) readCompressedDim(in store.IndexInput) (int, error) {
	compressedDim, err := in.ReadByte()
	if err != nil {
		return 0, err
	}

	config := r.config

	dim := int8(compressedDim)

	if dim < -2 || int(dim) >= config.numDims || (r.version < VERSION_LOW_CARDINALITY_LEAVES && dim == -2) {
		return 0, fmt.Errorf("got compressedDim=%d", dim)
	}
	return int(dim), nil
}

func (r *Reader) readCommonPrefixes(ctx context.Context, commonPrefixLengths []int, scratchPackedValue []byte, in store.IndexInput) error {
	config := r.config
	for dim := 0; dim < config.numDims; dim++ {
		prefix, err := in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		commonPrefixLengths[dim] = int(prefix)
		if prefix > 0 {
			fromIndex := dim * config.bytesPerDim
			toIndex := fromIndex + int(prefix)
			if _, err := in.Read(scratchPackedValue[fromIndex:toIndex]); err != nil {
				return err
			}
		}
	}
	return nil
}

// TODO: 检查intersect的实现
func (r *Reader) intersect(ctx context.Context, state *IntersectState, cellMinPacked, cellMaxPacked []byte) error {
	relation := state.visitor.Compare(cellMinPacked, cellMaxPacked)
	switch relation {
	case types.CELL_OUTSIDE_QUERY:
		return nil
	case types.CELL_INSIDE_QUERY:
		// This cell is fully inside of the query shape: recursively add all points in this cell without filtering
		return r.addAll(ctx, state, false)
		// The cell crosses the shape boundary, or the cell fully contains the query, so we fall through and do full filtering:
	}

	if state.index.IsLeafNode() {
		// TODO: we can assert that the first value here in fact matches what the index claimed?

		// In the unbalanced case it's possible the left most node only has one child:
		if state.index.NodeExists() {
			// Leaf node; scan and filter all points in this block:
			count, err := r.readDocIDs(ctx, state.in, state.index.GetLeafBlockFP(), state.scratchIterator)
			if err != nil {
				return err
			}
			// Again, this time reading values and checking with the visitor
			return r.visitDocValues(ctx, state.commonPrefixLengths, state.scratchDataPackedValue, state.scratchMinIndexPackedValue, state.scratchMaxIndexPackedValue, state.in, state.scratchIterator, count, state.visitor)
		}
	}

	config := r.config

	// Non-leaf node: recurse on the split left and right nodes
	splitDim := state.index.GetSplitDim()

	splitPackedValue := state.index.GetSplitPackedValue()
	splitDimValue := state.index.GetSplitDimValue()
	//System.out.println("  splitDimValue=" + splitDimValue + " splitDim=" + splitDim);

	// make sure cellMin <= splitValue <= cellMax:

	// Recurse on left sub-tree:
	arraycopy(cellMaxPacked, 0, splitPackedValue, 0, config.packedIndexBytesLength)
	arraycopy(splitDimValue, 0, splitPackedValue, splitDim*config.bytesPerDim, config.bytesPerDim)
	if err := state.index.PushLeft(ctx); err != nil {
		return err
	}
	if err := r.intersect(ctx, state, cellMinPacked, splitPackedValue); err != nil {
		return err
	}
	state.index.Pop()

	// Restore the split dim value since it may have been overwritten while recursing:
	arraycopy(splitPackedValue, splitDim*config.bytesPerDim, splitDimValue, 0, config.bytesPerDim)

	// Recurse on right sub-tree:
	arraycopy(cellMinPacked, 0, splitPackedValue, 0, config.packedIndexBytesLength)
	arraycopy(splitDimValue, 0, splitPackedValue, splitDim*config.bytesPerDim, config.bytesPerDim)
	if err := state.index.PushRight(ctx); err != nil {
		return err
	}

	if err := r.intersect(ctx, state, splitPackedValue, cellMaxPacked); err != nil {
		return err
	}
	state.index.Pop()

	return nil
}

func (r *Reader) estimatePointCount(ctx context.Context, state *IntersectState, cellMinPacked, cellMaxPacked []byte) (int, error) {
	relation := state.visitor.Compare(cellMinPacked, cellMaxPacked)

	config := r.config

	switch relation {
	case types.CELL_OUTSIDE_QUERY:
		return 0, nil
	case types.CELL_INSIDE_QUERY:
		return config.MaxPointsInLeafNode() * state.index.getNumLeaves(), nil
	default:
	}

	if state.index.IsLeafNode() {
		// Assume half the points matched
		return (config.MaxPointsInLeafNode() + 1) / 2, nil
	}

	// Non-leaf node: recurse on the split left and right nodes
	splitDim := state.index.GetSplitDim()
	splitPackedValue := state.index.GetSplitPackedValue()
	splitDimValue := state.index.GetSplitDimValue()

	//splitDimOffset := splitDim * config.bytesPerDim

	// Recurse on left sub-tree:
	arraycopy(cellMaxPacked, 0, splitPackedValue, 0, config.packedIndexBytesLength)
	arraycopy(splitDimValue, 0, splitPackedValue, splitDim*config.bytesPerDim, config.bytesPerDim)

	if err := state.index.PushLeft(ctx); err != nil {
		return 0, err
	}

	leftCost, err := r.estimatePointCount(ctx, state, cellMinPacked, splitPackedValue)
	if err != nil {
		return 0, err
	}
	state.index.Pop()

	// Restore the split dim value since it may have been overwritten while recursing:
	arraycopy(splitPackedValue, splitDim*config.bytesPerDim, splitDimValue, 0, config.bytesPerDim)

	// Recurse on right sub-tree:
	arraycopy(cellMinPacked, 0, splitPackedValue, 0, config.packedIndexBytesLength)
	arraycopy(splitDimValue, 0, splitPackedValue, splitDim*config.bytesPerDim, config.bytesPerDim)

	if err = state.index.PushRight(ctx); err != nil {
		return 0, err
	}
	rightCost, err := r.estimatePointCount(ctx, state, splitPackedValue, cellMaxPacked)
	state.index.Pop()
	return leftCost + rightCost, nil
}

func (r *Reader) EstimateDocCount(ctx context.Context, visitor types.IntersectVisitor) (int, error) {
	return types.EstimateDocCount(ctx, r, visitor)
}

func (r *Reader) GetMinPackedValue() ([]byte, error) {
	return slices.Clone(r.minPackedValue), nil
}

func (r *Reader) GetMaxPackedValue() ([]byte, error) {
	return slices.Clone(r.maxPackedValue), nil
}

func (r *Reader) GetNumDimensions() (int, error) {
	return r.config.numDims, nil
}

func (r *Reader) GetNumIndexDimensions() (int, error) {
	return r.config.numIndexDims, nil
}

func (r *Reader) GetBytesPerDimension() (int, error) {
	return r.config.bytesPerDim, nil
}

func (r *Reader) Size() int {
	return r.pointCount
}

func (r *Reader) GetDocCount() int {
	return r.docCount
}

func (r *Reader) IsLeafNode(nodeID int) bool {
	return nodeID >= r.leafNodeOffset
}

func (r *Reader) getMinLeafBlockFP() int64 {
	return r.minLeafBlockFP
}

func (r *Reader) getTreeDepth() int {
	// First +1 because all the non-leave nodes makes another power
	// of 2; e.g. to have a fully balanced tree with 4 leaves you
	// need a depth=3 tree:

	// Second +1 because MathUtil.log computes floor of the logarithm; e.g.
	// with 5 leaves you need a depth=4 tree:
	return int(math.Log2(float64(r.numLeaves))) + 2
}

func (r *Reader) addAll(ctx context.Context, state *IntersectState, grown bool) error {
	config := r.config

	if grown == false {
		maxPointCount := config.maxPointsInLeafNode * state.index.getNumLeaves()
		if maxPointCount <= math.MaxInt32 { // could be >MAX_VALUE if there are more than 2B points in total
			state.visitor.Grow(maxPointCount)
			grown = true
		}
	}

	if state.index.IsLeafNode() {

		if state.index.NodeExists() {
			err := r.visitDocIDs(ctx, state.in, state.index.GetLeafBlockFP(), state.visitor)
			if err != nil {
				return err
			}
		}
		// TODO: we can assert that the first value here in fact matches what the index claimed?
	} else {
		err := state.index.PushLeft(nil)
		if err != nil {
			return err
		}
		err = r.addAll(ctx, state, grown)
		if err != nil {
			return err
		}
		state.index.Pop()

		err = state.index.PushRight(nil)
		if err != nil {
			return err
		}
		err = r.addAll(ctx, state, grown)
		if err != nil {
			return err
		}
		state.index.Pop()
	}

	return nil
}
