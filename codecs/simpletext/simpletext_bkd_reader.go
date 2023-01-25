package simpletext

import (
	"bytes"
	"errors"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"io"
)

var _ index.PointValues = &SimpleTextBKDReader{}

type SimpleTextBKDReader struct {
	*TextReader

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
	pointCount             int64
	docCount               int
	version                int
	packedBytesLength      int
	packedIndexBytesLength int
}

func (s *SimpleTextBKDReader) Intersect(visitor index.IntersectVisitor) error {
	return s.intersect(s.getIntersectState(visitor), 1, s.minPackedValue, s.maxPackedValue)
}

// Fast path: this is called when the query box fully encompasses all cells under this node.
func (s *SimpleTextBKDReader) addAll(state *IntersectState, nodeID int) error {
	if nodeID >= s.leafNodeOffset {
		// TODO: we can assert that the first value here in fact matches what the index claimed?
		return s.visitDocIDs(state.in,
			s.leafBlockFPs[nodeID-s.leafNodeOffset], state.visitor)
	}

	if err := s.addAll(state, 2*nodeID); err != nil {
		return err
	}
	return s.addAll(state, 2*nodeID+1)
}

// Create a new SimpleTextBKDReader.IntersectState
func (s *SimpleTextBKDReader) getIntersectState(visitor index.IntersectVisitor) *IntersectState {
	return s.NewIntersectState(s.in.Clone(), s.numDims,
		s.packedBytesLength, s.maxPointsInLeafNode, visitor)
}

func (s *SimpleTextBKDReader) intersect(state *IntersectState, nodeID int, cellMinPacked, cellMaxPacked []byte) error {

	r := state.visitor.Compare(cellMinPacked, cellMaxPacked)
	switch r {
	case index.CELL_OUTSIDE_QUERY:
		// This cell is fully outside of the query shape: stop recursing
		return nil
	case index.CELL_INSIDE_QUERY:
		// This cell is fully inside of the query shape: recursively add all points in this cell without filtering
		return s.addAll(state, nodeID)
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
			if err := s.visitDocValues(state.commonPrefixLengths, state.scratchPackedValue,
				state.in, state.scratchDocIDs, count, state.visitor); err != nil {
				return err
			}

		}
		return nil
	}

	// Non-leaf node: recurse on the split left and right nodes
	address := nodeID * s.bytesPerIndexEntry
	splitDim := 0

	if s.numIndexDims != 1 {
		splitDim = int(s.splitPackedValues[address] & 0xff)
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

	if err := s.intersect(state, 2*nodeID, cellMinPacked, splitPackedValue); err != nil {
		return err
	}

	// Recurse on right sub-tree:
	copy(splitPackedValue, cellMinPacked)
	copy(splitPackedValue[splitDim*bytesPerDim:],
		s.splitPackedValues[address:address+bytesPerDim])
	if err := s.intersect(state, 2*nodeID+1, splitPackedValue, cellMaxPacked); err != nil {
		return err
	}
	return nil
}

func (s *SimpleTextBKDReader) visitDocIDs(in store.IndexInput, blockFP int64, visitor index.IntersectVisitor) error {
	scratch := new(bytes.Buffer)
	if _, err := in.Seek(blockFP, io.SeekStart); err != nil {
		return err
	}

	if err := ReadLine(in, scratch); err != nil {
		return err
	}

	count, err := ParseInt(scratch, BLOCK_COUNT)
	if err != nil {
		return err
	}

	visitor.Grow(count)
	for i := 0; i < count; i++ {
		if err := ReadLine(in, scratch); err != nil {
			return err
		}
		docID, err := ParseInt(scratch, BLOCK_DOC_ID)
		if err != nil {
			return err
		}

		if err := visitor.VisitByDocID(docID); err != nil {
			return err
		}
	}
	return nil
}

func (s *SimpleTextBKDReader) readDocIDs(in store.IndexInput, blockFP int64, docIDs []int) (int, error) {
	scratch := new(bytes.Buffer)
	if _, err := in.Seek(blockFP, io.SeekStart); err != nil {
		return 0, err
	}

	if err := ReadLine(in, scratch); err != nil {
		return 0, err
	}

	count, err := ParseInt(scratch, BLOCK_COUNT)
	if err != nil {
		return 0, err
	}

	for idx := 0; idx < count; idx++ {
		if err := ReadLine(in, scratch); err != nil {
			return 0, err
		}
		docID, err := ParseInt(scratch, BLOCK_DOC_ID)
		if err != nil {
			return 0, err
		}
		docIDs[idx] = docID
	}
	return count, nil
}

func (s *SimpleTextBKDReader) visitDocValues(commonPrefixLengths []int, scratchPackedValue []byte,
	in store.IndexInput, docIDs []int, count int, visitor index.IntersectVisitor) error {

	visitor.Grow(count)
	// NOTE: we don't do prefix coding, so we ignore commonPrefixLengths
	// assert scratchPackedValue.length == packedBytesLength;

	scratch := new(bytes.Buffer)
	for idx := 0; idx < count; idx++ {
		if err := ReadLine(in, scratch); err != nil {
			return err
		}

		if !bytes.HasPrefix(scratch.Bytes(), BLOCK_VALUE) {
			return errors.New("prefix is not block value")
		}

		scratch.Next(len(BLOCK_VALUE))

		value := util.BytesToString(scratch.Bytes())
		br := []byte(value)

		packedBytesLength := s.packedBytesLength

		if len(br) != packedBytesLength {
			return errors.New("size of value not equal")
		}

		copy(scratchPackedValue, br)
		if err := visitor.Visit(docIDs[idx], scratchPackedValue); err != nil {
			return err
		}
	}
	return nil
}

func (s *SimpleTextBKDReader) EstimatePointCount(visitor index.IntersectVisitor) int64 {
	return s.estimatePointCount(s.getIntersectState(visitor), 1, s.minPackedValue, s.maxPackedValue)
}

func (s *SimpleTextBKDReader) estimatePointCount(state *IntersectState, nodeID int,
	cellMinPacked, cellMaxPacked []byte) int64 {

	r := state.visitor.Compare(cellMinPacked, cellMaxPacked)
	if r == index.CELL_OUTSIDE_QUERY {
		return 0
	} else if nodeID >= s.leafNodeOffset {
		return int64(s.maxPointsInLeafNode)
	}

	// Non-leaf node: recurse on the split left and right nodes
	address := nodeID * s.bytesPerIndexEntry
	splitDim := 0
	if s.numIndexDims != 1 {
		splitDim = int(s.splitPackedValues[address] & 0xff)
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

func (s *SimpleTextBKDReader) GetMinPackedValue() ([]byte, error) {
	bs := make([]byte, len(s.minPackedValue))
	copy(bs, s.minPackedValue)
	return bs, nil
}

func (s *SimpleTextBKDReader) GetMaxPackedValue() ([]byte, error) {
	bs := make([]byte, len(s.maxPackedValue))
	copy(bs, s.maxPackedValue)
	return bs, nil
}

func (s *SimpleTextBKDReader) GetNumDimensions() (int, error) {
	return s.numDims, nil
}

func (s *SimpleTextBKDReader) GetNumIndexDimensions() (int, error) {
	return s.numIndexDims, nil
}

func (s *SimpleTextBKDReader) GetBytesPerDimension() (int, error) {
	return s.bytesPerDim, nil
}

func (s *SimpleTextBKDReader) Size() int64 {
	return s.pointCount
}

func (s *SimpleTextBKDReader) GetDocCount() int {
	return s.docCount
}

// IntersectState Used to track all state for a single call to intersect.
type IntersectState struct {
	reader *SimpleTextBKDReader

	in                  store.IndexInput
	scratchDocIDs       []int
	scratchPackedValue  []byte
	commonPrefixLengths []int
	visitor             index.IntersectVisitor
}

func (s *SimpleTextBKDReader) NewIntersectState(in store.IndexInput,
	numDims, packedBytesLength, maxPointsInLeafNode int,
	visitor index.IntersectVisitor) *IntersectState {

	return &IntersectState{
		reader:              s,
		in:                  in,
		scratchDocIDs:       make([]int, numDims),
		scratchPackedValue:  make([]byte, maxPointsInLeafNode),
		commonPrefixLengths: make([]int, packedBytesLength),
		visitor:             visitor,
	}
}
