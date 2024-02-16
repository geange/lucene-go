package bkd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/types"
)

var _ types.IntersectVisitor = &VerifyPointsVisitor{}

// VerifyPointsVisitor
// Walks the entire N-dimensional points space, verifying that all points fall within the last cell's boundaries.
// lucene.internal
type VerifyPointsVisitor struct {
	pointCountSeen        int64
	lastDocID             int
	maxDoc                int
	docsSeen              *bitset.BitSet
	lastMinPackedValue    []byte
	lastMaxPackedValue    []byte
	lastPackedValue       []byte
	globalMinPackedValue  []byte
	globalMaxPackedValue  []byte
	packedBytesCount      int
	packedIndexBytesCount int
	numDataDims           int
	numIndexDims          int
	bytesPerDim           int
	fieldName             string
}

func NewVerifyPointsVisitor(fieldName string, maxDoc int, values types.PointValues) (*VerifyPointsVisitor, error) {
	numDataDims, err := values.GetNumDimensions()
	if err != nil {
		return nil, err
	}
	numIndexDims, err := values.GetNumIndexDimensions()
	if err != nil {
		return nil, err
	}
	bytesPerDim, err := values.GetBytesPerDimension()
	if err != nil {
		return nil, err
	}
	packedBytesCount := numDataDims * bytesPerDim
	packedIndexBytesCount := numIndexDims * bytesPerDim
	globalMinPackedValue, err := values.GetMinPackedValue()
	if err != nil {
		return nil, err
	}
	globalMaxPackedValue, err := values.GetMaxPackedValue()
	if err != nil {
		return nil, err
	}
	docsSeen := bitset.New(uint(maxDoc))
	lastMinPackedValue := make([]byte, packedIndexBytesCount)
	lastMaxPackedValue := make([]byte, packedIndexBytesCount)
	lastPackedValue := make([]byte, packedBytesCount)

	return &VerifyPointsVisitor{
		pointCountSeen:        0,
		lastDocID:             -1,
		maxDoc:                maxDoc,
		docsSeen:              docsSeen,
		lastMinPackedValue:    lastMinPackedValue,
		lastMaxPackedValue:    lastMaxPackedValue,
		lastPackedValue:       lastPackedValue,
		globalMinPackedValue:  globalMinPackedValue,
		globalMaxPackedValue:  globalMaxPackedValue,
		packedBytesCount:      packedBytesCount,
		packedIndexBytesCount: packedIndexBytesCount,
		numDataDims:           numDataDims,
		numIndexDims:          numIndexDims,
		bytesPerDim:           bytesPerDim,
		fieldName:             fieldName,
	}, nil
}

func (v *VerifyPointsVisitor) Visit(ctx context.Context, docID int) error {
	return errors.New("not available")
}

func (v *VerifyPointsVisitor) VisitLeaf(ctx context.Context, docID int, packedValue []byte) error {
	v.pointCountSeen++
	v.docsSeen.Set(uint(docID))

	for dim := 0; dim < v.numIndexDims; dim++ {
		fromIndex := v.bytesPerDim * dim
		toIndex := fromIndex + v.bytesPerDim

		// Compare to last cell:
		if bytes.Compare(packedValue[fromIndex:toIndex], v.lastMinPackedValue[fromIndex:toIndex]) < 0 {
			// This doc's point, in this dimension, is lower than the minimum value of the last cell checked:
			return fmt.Errorf("docId=%d, in this dimension, is lower than the minimum value of the last cell checked", docID)
		}

		if bytes.Compare(packedValue[fromIndex:toIndex], v.lastMaxPackedValue[fromIndex:toIndex]) > 0 {
			// This doc's point, in this dimension, is greater than the maximum value of the last cell checked:
			return fmt.Errorf("docId=%d, in this dimension, is greater than the maximum value of the last cell checked", docID)
		}
	}

	// In the 1D data case, PointValues must make a single in-order sweep through all values, and tie-break by
	// increasing docID:
	// for data dimension > 1, leaves are sorted by the dimension with the lowest cardinality to improve block compression
	if v.numDataDims == 1 {
		cmp := bytes.Compare(v.lastPackedValue[:v.bytesPerDim], packedValue[:v.bytesPerDim])
		if cmp >= 0 {
			return errors.New("last packed value bigger")
		}
		copy(v.lastPackedValue, packedValue[:v.bytesPerDim])
		v.lastDocID = docID
	}

	return nil
}

func (v *VerifyPointsVisitor) Compare(minPackedValue, maxPackedValue []byte) types.Relation {
	arraycopy(minPackedValue, 0, v.lastMinPackedValue, 0, v.packedIndexBytesCount)
	arraycopy(maxPackedValue, 0, v.lastMaxPackedValue, 0, v.packedIndexBytesCount)

	bytesPerDim := v.bytesPerDim
	globalMinPackedValue := v.globalMinPackedValue
	globalMaxPackedValue := v.globalMaxPackedValue

	for dim := 0; dim < v.numIndexDims; dim++ {
		fromIndex := v.bytesPerDim * dim
		toIndex := fromIndex + bytesPerDim

		if compareUnsigned(minPackedValue, fromIndex, toIndex, maxPackedValue, fromIndex, toIndex) > 0 {
			panic("CheckIndexException")
		}

		// Make sure this cell is not outside of the global min/max:
		if compareUnsigned(minPackedValue, fromIndex, toIndex, globalMinPackedValue, fromIndex, toIndex) < 0 {
			panic("CheckIndexException")
		}

		if compareUnsigned(maxPackedValue, fromIndex, toIndex, globalMinPackedValue, fromIndex, toIndex) < 0 {
			panic("CheckIndexException")
		}

		if compareUnsigned(minPackedValue, fromIndex, toIndex, globalMaxPackedValue, fromIndex, toIndex) > 0 {
			panic("CheckIndexException")
		}

		if compareUnsigned(maxPackedValue, fromIndex, toIndex, globalMaxPackedValue, fromIndex, toIndex) > 0 {
			panic("CheckIndexException")
		}

		if compareUnsigned(maxPackedValue, fromIndex, toIndex, globalMinPackedValue, fromIndex, toIndex) < 0 {
			panic("CheckIndexException")
		}

		if compareUnsigned(minPackedValue, fromIndex, toIndex, globalMaxPackedValue, fromIndex, toIndex) > 0 {
			panic("CheckIndexException")
		}

		if compareUnsigned(maxPackedValue, fromIndex, toIndex, globalMaxPackedValue, fromIndex, toIndex) > 0 {
			panic("CheckIndexException")
		}
	}

	// We always pretend the query shape is so complex that it crosses every cell, so
	// that packedValue is passed for every document
	return types.CELL_CROSSES_QUERY
}

func (v *VerifyPointsVisitor) Grow(count int) {
}
