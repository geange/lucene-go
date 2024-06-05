package bkd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/geange/lucene-go/core/store"
)

type OneDimensionBKDWriter struct {
	p                          *Writer
	metaOut, indexOut, dataOut store.IndexOutput
	dataStartFP                int64
	leafBlockFPs               []int64
	leafBlockStartValues       [][]byte
	leafValues                 []byte
	leafDocs                   []int
	valueCount                 int
	leafCount                  int
	leafCardinality            int

	lastPackedValue []byte
	lastDocID       int
}

func (w *Writer) NewOneDimensionBKDWriter(metaOut, indexOut, dataOut store.IndexOutput) (*OneDimensionBKDWriter, error) {
	if w.config.NumIndexDims() != 1 {
		return nil, errors.New("config.numIndexDims must be 1")
	}
	if w.pointCount != 0 {
		return nil, errors.New("cannot mix add and merge")
	}

	// Catch user silliness:
	if w.finished.Load() == true {
		return nil, errors.New("already finished")
	}

	// Mark that we already finished:
	w.finished.Store(false)

	writer := &OneDimensionBKDWriter{
		p:                    w,
		metaOut:              metaOut,
		indexOut:             indexOut,
		dataOut:              dataOut,
		dataStartFP:          dataOut.GetFilePointer(),
		leafBlockFPs:         make([]int64, 0),
		leafBlockStartValues: make([][]byte, 0),
		leafValues:           make([]byte, w.config.MaxPointsInLeafNode()*w.config.PackedBytesLength()),
		leafDocs:             make([]int, w.config.MaxPointsInLeafNode()),
		lastPackedValue:      make([]byte, w.config.PackedBytesLength()),
	}

	return writer, nil
}

func (r *OneDimensionBKDWriter) Add(ctx context.Context, packedValue []byte, docID int) error {
	w := r.p
	config := w.config

	bytesPerDim := config.BytesPerDim()
	from := (r.leafCount - 1) * bytesPerDim
	to := from + bytesPerDim

	if r.leafCount == 0 ||
		bytes.Equal(r.leafValues[from:to], packedValue[:bytesPerDim]) == false {
		r.leafCardinality++
	}

	packedBytesLength := config.PackedBytesLength()
	destFrom := r.leafCount * config.PackedBytesLength()
	destTo := destFrom + packedBytesLength
	copy(r.leafValues[destFrom:destTo], packedValue[:packedBytesLength])

	r.leafDocs[r.leafCount] = docID
	w.docsSeen.Set(uint(docID))
	r.leafCount++

	if r.valueCount+r.leafCount > w.totalPointCount {
		return fmt.Errorf("totalPointCount=%d was passed when we were created", w.totalPointCount)
	}

	if r.leafCount == config.maxPointsInLeafNode {
		// We write a block once we hit exactly the max count ... this is different from
		// when we write N > 1 dimensional points where we write between max/2 and max per leaf block
		if err := r.writeLeafBlock(nil, r.leafCardinality); err != nil {
			return err
		}
		r.leafCardinality = 0
		r.leafCount = 0
	}
	return nil
}

func (r *OneDimensionBKDWriter) Finish() (Runnable, error) {
	w := r.p
	config := w.config

	if r.leafCount > 0 {
		if err := r.writeLeafBlock(nil, r.leafCardinality); err != nil {
			return nil, err
		}
		r.leafCardinality = 0
		r.leafCount = 0
	}

	if r.valueCount == 0 {
		return emptyRunnable, nil
	}

	leafNodes := &oneDimBKDTreeLeafNodes{dimWriter: r}

	return func(ctx context.Context) error {
		return w.writeIndex(ctx, r.metaOut, r.indexOut, config.MaxPointsInLeafNode(), leafNodes, r.dataStartFP)
	}, nil
}

var _ LeafNodes = &oneDimBKDTreeLeafNodes{}

type oneDimBKDTreeLeafNodes struct {
	dimWriter *OneDimensionBKDWriter
}

func (r *oneDimBKDTreeLeafNodes) NumLeaves() int {
	return len(r.dimWriter.leafBlockFPs)
}

func (r *oneDimBKDTreeLeafNodes) GetLeafLP(index int) int64 {
	return r.dimWriter.leafBlockFPs[index]
}

func (r *oneDimBKDTreeLeafNodes) GetSplitValue(index int) []byte {
	return r.dimWriter.leafBlockStartValues[index]
}

func (r *oneDimBKDTreeLeafNodes) GetSplitDimension(index int) int {
	return 0
}

func (r *OneDimensionBKDWriter) writeLeafBlock(ctx context.Context, leafCardinality int) error {
	w := r.p
	conf := w.config

	packedIndexBytesLength := conf.PackedIndexBytesLength()
	if r.valueCount == 0 {
		copy(w.minPackedValue[:packedIndexBytesLength], r.leafValues[:packedIndexBytesLength])
	}
	srcFrom := (r.leafCount - 1) * conf.PackedBytesLength()
	srcTo := srcFrom + packedIndexBytesLength
	dstFrom := 0
	dstTo := dstFrom + packedIndexBytesLength
	copy(w.maxPackedValue[dstFrom:dstTo], r.leafValues[srcFrom:srcTo])

	r.valueCount += r.leafCount

	if len(r.leafBlockFPs) > 0 {
		// Save the first (minimum) value in each leaf block except the first, to build the split value index in the end:
		tmpValues := slices.Clone(r.leafValues[:conf.PackedBytesLength()])
		r.leafBlockStartValues = append(r.leafBlockStartValues, tmpValues)
	}
	r.leafBlockFPs = append(r.leafBlockFPs, r.dataOut.GetFilePointer())

	// Find per-dim common prefix:
	offset := (r.leafCount - 1) * conf.PackedBytesLength()
	prefix := Mismatch(r.leafValues[0:conf.bytesPerDim], r.leafValues[offset:offset+conf.BytesPerDim()])
	if prefix == -1 {
		prefix = conf.BytesPerDim()
	}

	w.commonPrefixLengths[0] = prefix

	err := w.writeLeafBlockDocs(w.scratchOut, r.leafDocs[0:r.leafCount])
	if err != nil {
		return err
	}
	err = w.writeCommonPrefixes(ctx, w.scratchOut, w.commonPrefixLengths, r.leafValues)
	if err != nil {
		return err
	}

	packedValues := func(i int) []byte {
		from := conf.PackedBytesLength() * i
		to := from + conf.PackedBytesLength()
		return r.leafValues[from:to]
	}

	err = w.writeLeafBlockPackedValues(w.scratchOut, w.commonPrefixLengths, r.leafCount,
		0, packedValues, leafCardinality)
	if err != nil {
		return err
	}
	err = w.scratchOut.CopyTo(r.dataOut)
	if err != nil {
		return err
	}
	w.scratchOut.Reset()
	return nil
}

func (r *OneDimensionBKDWriter) checkMaxLeafNodeCount(numLeaves int) error {
	return nil
}
