package simpletext

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/bytesref"
)

var _ types.PointValues = &BKDReader{}

type BKDReader struct {
	splitPackedValues      []byte
	leafBlockFPs           []int64
	leafNodeOffset         int
	numDims                int
	numIndexDims           int
	bytesPerDim            int
	bytesPerIndexEntry     int
	in                     store.IndexInput
	maxPointsInLeafNode    int
	minPackedValue         []byte
	maxPackedValue         []byte
	pointCount             int
	docCount               int
	version                int
	packedBytesLength      int
	packedIndexBytesLength int
}

func NewBKDReader(in store.IndexInput, numDims, numIndexDims, maxPointsInLeafNode, bytesPerDim int,
	leafBlockFPs []int64, splitPackedValues []byte, minPackedValue, maxPackedValue []byte,
	pointCount int, docCount int) (*BKDReader, error) {

	bytesPerIndexEntry := bytesPerDim + 1
	if numIndexDims == 1 {
		bytesPerIndexEntry = bytesPerDim
	}

	reader := &BKDReader{
		splitPackedValues:      splitPackedValues,
		leafBlockFPs:           leafBlockFPs,
		leafNodeOffset:         len(leafBlockFPs),
		numDims:                numDims,
		numIndexDims:           numIndexDims,
		bytesPerDim:            bytesPerDim,
		bytesPerIndexEntry:     bytesPerIndexEntry,
		in:                     in,
		maxPointsInLeafNode:    maxPointsInLeafNode,
		minPackedValue:         minPackedValue,
		maxPackedValue:         maxPackedValue,
		pointCount:             pointCount,
		docCount:               docCount,
		version:                VERSION_CURRENT,
		packedBytesLength:      numDims * bytesPerDim,
		packedIndexBytesLength: numIndexDims * bytesPerDim,
	}

	if len(minPackedValue) != reader.packedIndexBytesLength {
		return nil, errors.New("len(minPackedValue) != reader.packedIndexBytesLength")
	}

	if len(maxPackedValue) != reader.packedIndexBytesLength {
		return nil, errors.New("len(maxPackedValue) != reader.packedIndexBytesLength")
	}

	return reader, nil
}

func (s *BKDReader) Intersect(ctx context.Context, visitor types.IntersectVisitor) error {
	return s.intersect(nil, s.getIntersectState(visitor), 1, s.minPackedValue, s.maxPackedValue)
}

// Fast path: this is called when the query box fully encompasses all cells under this node.
func (s *BKDReader) addAll(ctx context.Context, state *IntersectState, nodeID int) error {
	if nodeID >= s.leafNodeOffset {
		// TODO: we can assert that the first value here in fact matches what the index claimed?
		return s.visitDocIDs(ctx, state.in, s.leafBlockFPs[nodeID-s.leafNodeOffset], state.visitor)
	}

	if err := s.addAll(ctx, state, 2*nodeID); err != nil {
		return err
	}
	return s.addAll(ctx, state, 2*nodeID+1)
}

// Create a new SimpleTextBKDReader.IntersectState
func (s *BKDReader) getIntersectState(visitor types.IntersectVisitor) *IntersectState {
	return s.NewIntersectState(s.in.Clone().(store.IndexInput), s.numDims,
		s.packedBytesLength, s.maxPointsInLeafNode, visitor)
}

func (s *BKDReader) intersect(ctx context.Context, state *IntersectState, nodeID int, cellMinPacked, cellMaxPacked []byte) error {

	r := state.visitor.Compare(cellMinPacked, cellMaxPacked)
	switch r {
	case types.CELL_OUTSIDE_QUERY:
		// This cell is fully outside of the query shape: stop recursing
		return nil
	case types.CELL_INSIDE_QUERY:
		// This cell is fully inside of the query shape: recursively add all points in this cell without filtering
		return s.addAll(nil, state, nodeID)
	default:
		// The cell crosses the shape boundary, or the cell fully contains the query,
		// so we fall through and do full filtering
	}

	if nodeID >= s.leafNodeOffset {
		// TODO: we can assert that the first value here in fact matches what the index claimed?

		leafID := nodeID - s.leafNodeOffset

		// In the unbalanced case it's possible the left most node only has one child:
		if leafID < len(s.leafBlockFPs) {
			// Leaf node; scan and filter all points in this block:
			count, err := s.readDocIDs(state.in, s.leafBlockFPs[leafID], state.scratchDocIDs)
			if err != nil {
				return err
			}

			// Again, this time reading values and checking with the visitor
			if err := s.visitDocValues(ctx, state.commonPrefixLengths, state.scratchPackedValue, state.in, state.scratchDocIDs, count, state.visitor); err != nil {
				return err
			}
		}
		return nil
	}

	// Non-leaf node: recurse on the split left and right nodes
	address := nodeID * s.bytesPerIndexEntry
	splitDim := 0

	if s.numIndexDims != 1 {
		splitDim = int(s.splitPackedValues[address])
		address++
	}

	//assert splitDim < numIndexDims;

	// TODO: can we alloc & reuse this up front?

	bytesPerDim := s.bytesPerDim
	packedIndexBytesLength := s.packedIndexBytesLength

	splitPackedValue := make([]byte, packedIndexBytesLength)

	// Recurse on left sub-tree:
	copy(splitPackedValue, cellMaxPacked)
	copy(splitPackedValue[splitDim*bytesPerDim:],
		s.splitPackedValues[address:address+bytesPerDim])

	if err := s.intersect(ctx, state, 2*nodeID, cellMinPacked, splitPackedValue); err != nil {
		return err
	}

	// Recurse on right sub-tree:
	copy(splitPackedValue, cellMinPacked)
	copy(splitPackedValue[splitDim*bytesPerDim:],
		s.splitPackedValues[address:address+bytesPerDim])
	if err := s.intersect(nil, state, 2*nodeID+1, splitPackedValue, cellMaxPacked); err != nil {
		return err
	}
	return nil
}

func (s *BKDReader) visitDocIDs(ctx context.Context, in store.IndexInput, blockFP int64, visitor types.IntersectVisitor) error {
	scratch := new(bytes.Buffer)
	if _, err := in.Seek(blockFP, io.SeekStart); err != nil {
		return err
	}

	if err := utils.ReadLine(in, scratch); err != nil {
		return err
	}

	count, err := utils.ParseInt(scratch, BLOCK_COUNT)
	if err != nil {
		return err
	}

	visitor.Grow(count)
	for i := 0; i < count; i++ {
		if err := utils.ReadLine(in, scratch); err != nil {
			return err
		}
		docID, err := utils.ParseInt(scratch, BLOCK_DOC_ID)
		if err != nil {
			return err
		}

		if err := visitor.Visit(ctx, docID); err != nil {
			return err
		}
	}
	return nil
}

func (s *BKDReader) readDocIDs(in store.IndexInput, blockFP int64, docIDs []int) (int, error) {
	scratch := new(bytes.Buffer)
	if _, err := in.Seek(blockFP, io.SeekStart); err != nil {
		return 0, err
	}

	if err := utils.ReadLine(in, scratch); err != nil {
		return 0, err
	}

	count, err := utils.ParseInt(scratch, BLOCK_COUNT)
	if err != nil {
		return 0, err
	}

	for idx := 0; idx < count; idx++ {
		if err := utils.ReadLine(in, scratch); err != nil {
			return 0, err
		}
		docID, err := utils.ParseInt(scratch, BLOCK_DOC_ID)
		if err != nil {
			return 0, err
		}
		docIDs[idx] = docID
	}
	return count, nil
}

func (s *BKDReader) visitDocValues(ctx context.Context, commonPrefixLengths []int, scratchPackedValue []byte, in store.IndexInput, docIDs []int, count int, visitor types.IntersectVisitor) error {

	visitor.Grow(count)
	// NOTE: we don't do prefix coding, so we ignore commonPrefixLengths
	// assert scratchPackedValue.length == packedBytesLength;

	scratch := new(bytes.Buffer)
	for idx := 0; idx < count; idx++ {
		if err := utils.ReadLine(in, scratch); err != nil {
			return err
		}

		if !bytes.HasPrefix(scratch.Bytes(), BLOCK_VALUE) {
			return errors.New("prefix is not block value")
		}

		scratch.Next(len(BLOCK_VALUE))

		//value := util.BytesToString(scratch.NewBytes())
		//br := []byte(value)
		br, err := bytesref.StringToBytes(scratch.String())
		if err != nil {
			return err
		}

		packedBytesLength := s.packedBytesLength

		if len(br) != packedBytesLength {
			return errors.New("size of value not equal")
		}

		copy(scratchPackedValue, br)
		if err := visitor.VisitLeaf(ctx, docIDs[idx], scratchPackedValue); err != nil {
			return err
		}
	}
	return nil
}

func (s *BKDReader) EstimatePointCount(ctx context.Context, visitor types.IntersectVisitor) (int, error) {
	return s.estimatePointCount(s.getIntersectState(visitor), 1, s.minPackedValue, s.maxPackedValue), nil
}

func (s *BKDReader) EstimateDocCount(visitor types.IntersectVisitor) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *BKDReader) estimatePointCount(state *IntersectState, nodeID int,
	cellMinPacked, cellMaxPacked []byte) int {

	r := state.visitor.Compare(cellMinPacked, cellMaxPacked)
	if r == types.CELL_OUTSIDE_QUERY {
		return 0
	} else if nodeID >= s.leafNodeOffset {
		return s.maxPointsInLeafNode
	}

	// Non-leaf node: recurse on the split left and right nodes
	address := nodeID * s.bytesPerIndexEntry
	splitDim := 0
	if s.numIndexDims != 1 {
		splitDim = int(s.splitPackedValues[address])
		address++
	}

	// assert splitDim < numIndexDims;

	// TODO: can we alloc & reuse this up front?

	splitPackedValue := make([]byte, s.packedIndexBytesLength)

	bytesPerDim := s.bytesPerDim

	// Recurse on left sub-tree:
	copy(splitPackedValue, cellMaxPacked)
	copy(splitPackedValue[splitDim*bytesPerDim:], s.splitPackedValues[address:address+bytesPerDim])
	leftCost := s.estimatePointCount(state, 2*nodeID, cellMinPacked, splitPackedValue)

	// Recurse on right sub-tree:
	copy(splitPackedValue, cellMinPacked)
	copy(splitPackedValue[splitDim*bytesPerDim:], s.splitPackedValues[address:address+bytesPerDim])
	rightCost := s.estimatePointCount(state, 2*nodeID+1, splitPackedValue, cellMaxPacked)
	return leftCost + rightCost
}

func (s *BKDReader) GetMinPackedValue() ([]byte, error) {
	bs := make([]byte, len(s.minPackedValue))
	copy(bs, s.minPackedValue)
	return bs, nil
}

func (s *BKDReader) GetMaxPackedValue() ([]byte, error) {
	bs := make([]byte, len(s.maxPackedValue))
	copy(bs, s.maxPackedValue)
	return bs, nil
}

func (s *BKDReader) GetNumDimensions() (int, error) {
	return s.numDims, nil
}

func (s *BKDReader) GetNumIndexDimensions() (int, error) {
	return s.numIndexDims, nil
}

func (s *BKDReader) GetBytesPerDimension() (int, error) {
	return s.bytesPerDim, nil
}

func (s *BKDReader) Size() int {
	return s.pointCount
}

func (s *BKDReader) GetDocCount() int {
	return s.docCount
}

// IntersectState Used to track all state for a single call to intersect.
type IntersectState struct {
	reader              *BKDReader
	in                  store.IndexInput
	scratchDocIDs       []int
	scratchPackedValue  []byte
	commonPrefixLengths []int
	visitor             types.IntersectVisitor
}

func (s *BKDReader) NewIntersectState(in store.IndexInput,
	numDims, packedBytesLength, maxPointsInLeafNode int,
	visitor types.IntersectVisitor) *IntersectState {

	return &IntersectState{
		reader:              s,
		in:                  in,
		scratchDocIDs:       make([]int, maxPointsInLeafNode),
		scratchPackedValue:  make([]byte, packedBytesLength),
		commonPrefixLengths: make([]int, numDims),
		visitor:             visitor,
	}
}
