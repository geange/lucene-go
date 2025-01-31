package simpletext

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/bytesref"
)

var _ index.PointsReader = &PointsReader{}

type PointsReader struct {
	dataIn    store.IndexInput
	readState *index.SegmentReadState
	readers   map[string]*BKDReader
	scratch   *bytes.Buffer
}

func NewPointsReader(ctx context.Context, readState *index.SegmentReadState) (*PointsReader, error) {
	fieldToFileOffset := make(map[string]int64)
	indexFileName := store.SegmentFileName(readState.SegmentInfo.Name(), readState.SegmentSuffix, POINT_INDEX_EXTENSION)
	input, err := store.OpenChecksumInput(ctx, readState.Directory, indexFileName)
	if err != nil {
		return nil, err
	}

	reader := &PointsReader{
		dataIn:    nil,
		readState: readState,
		readers:   map[string]*BKDReader{},
		scratch:   new(bytes.Buffer),
	}

	tReader := utils.NewTextReader(input, reader.scratch)
	count, err := tReader.ParseInt(FIELD_COUNT)
	if err != nil {
		return nil, err
	}

	for i := 0; i < count; i++ {
		fieldName, err := tReader.ReadLabel(FIELD_FP_NAME)
		if err != nil {
			return nil, err
		}

		fp, err := tReader.ParseInt64(FIELD_FP)
		if err != nil {
			return nil, err
		}
		fieldToFileOffset[fieldName] = fp
	}
	if err := utils.CheckFooter(input); err != nil {
		return nil, err
	}

	fileName := store.SegmentFileName(readState.SegmentInfo.Name(), readState.SegmentSuffix, POINT_EXTENSION)
	reader.dataIn, err = readState.Directory.OpenInput(ctx, fileName)
	if err != nil {
		return nil, err
	}

	for k, v := range fieldToFileOffset {
		initReader, err := reader.initReader(v)
		if err != nil {
			return nil, err
		}
		reader.readers[k] = initReader
	}
	return reader, nil
}

func (s *PointsReader) Close() error {
	return s.dataIn.Close()
}

func (s *PointsReader) CheckIntegrity() error {
	// TODO: impl it
	return nil
}

func (s *PointsReader) GetValues(ctx context.Context, field string) (types.PointValues, error) {
	fieldInfo := s.readState.FieldInfos.FieldInfo(field)
	if fieldInfo == nil {
		return nil, fmt.Errorf("field=%s is unrecognized", field)
	}

	if fieldInfo.GetPointDimensionCount() == 0 {
		return nil, fmt.Errorf("field=%s did not index points", field)
	}
	return s.readers[field], nil
}

func (s *PointsReader) GetMergeInstance() index.PointsReader {
	return s
}

func (s *PointsReader) initReader(fp int64) (*BKDReader, error) {
	// NOTE: matches what writeIndex does in SimpleTextPointsWriter
	if _, err := s.dataIn.Seek(fp, io.SeekStart); err != nil {
		return nil, err
	}

	tr := utils.NewTextReader(s.dataIn, s.scratch)
	numDataDims, err := tr.ParseInt(NUM_DATA_DIMS)
	if err != nil {
		return nil, err
	}

	numIndexDims, err := tr.ParseInt(NUM_INDEX_DIMS)
	if err != nil {
		return nil, err
	}

	bytesPerDim, err := tr.ParseInt(BYTES_PER_DIM)
	if err != nil {
		return nil, err
	}

	maxPointsInLeafNode, err := tr.ParseInt(MAX_LEAF_POINTS)
	if err != nil {
		return nil, err
	}

	count, err := tr.ParseInt(INDEX_COUNT)
	if err != nil {
		return nil, err
	}

	v, err := tr.ReadLabel(MIN_VALUE)
	if err != nil {
		return nil, err
	}
	minValue, err := bytesref.StringToBytes(v)
	if err != nil {
		return nil, err
	}

	v, err = tr.ReadLabel(MAX_VALUE)
	if err != nil {
		return nil, err
	}
	maxValue, err := bytesref.StringToBytes(v)
	if err != nil {
		return nil, err
	}

	pointCount, err := tr.ParseInt(POINT_COUNT)
	if err != nil {
		return nil, err
	}

	docCount, err := tr.ParseInt(DOC_COUNT)
	if err != nil {
		return nil, err
	}

	leafBlockFPs := make([]int64, 0, count)
	for i := 0; i < count; i++ {
		blockFP, err := tr.ParseInt64(BLOCK_FP)
		if err != nil {
			return nil, err
		}
		leafBlockFPs = append(leafBlockFPs, blockFP)
	}

	count, err = tr.ParseInt(SPLIT_COUNT)
	if err != nil {
		return nil, err
	}

	bytesPerIndexEntry := bytesPerDim + 1
	if numIndexDims == 1 {
		bytesPerIndexEntry = bytesPerDim
	}
	splitPackedValues := make([]byte, count*bytesPerIndexEntry)

	for i := 0; i < count; i++ {
		address := bytesPerIndexEntry * i
		splitDim, err := tr.ParseInt(SPLIT_DIM)
		if err != nil {
			return nil, err
		}

		if numIndexDims != 1 {
			splitPackedValues[address] = byte(splitDim)
			address++
		}

		splitValue, err := tr.ReadLabel(SPLIT_VALUE)
		if err != nil {
			return nil, err
		}
		br, err := bytesref.StringToBytes(splitValue)
		if err != nil {
			return nil, err
		}
		copy(splitPackedValues[address:], br)
	}

	return NewBKDReader(s.dataIn, numDataDims, numIndexDims, maxPointsInLeafNode,
		bytesPerDim, leafBlockFPs, splitPackedValues, minValue, maxValue, pointCount, docCount)
}
