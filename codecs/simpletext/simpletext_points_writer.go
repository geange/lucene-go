package simpletext

import (
	"bytes"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

var _ index.PointsWriter = &SimpleTextPointsWriter{}

var (
	NUM_DATA_DIMS   = []byte("num data dims ")
	NUM_INDEX_DIMS  = []byte("num index dims ")
	BYTES_PER_DIM   = []byte("bytes per dim ")
	MAX_LEAF_POINTS = []byte("max leaf points ")
	INDEX_COUNT     = []byte("index count ")
	BLOCK_COUNT     = []byte("block count ")
	BLOCK_DOC_ID    = []byte("  doc ")
	BLOCK_FP        = []byte("  block fp ")
	BLOCK_VALUE     = []byte("  block value ")
	SPLIT_COUNT     = []byte("split count ")
	SPLIT_DIM       = []byte("  split dim ")
	SPLIT_VALUE     = []byte("  split value ")
	FIELD_COUNT     = []byte("field count ")
	FIELD_FP_NAME   = []byte("  field fp name ")
	FIELD_FP        = []byte("  field fp ")
	MIN_VALUE       = []byte("min value ")
	MAX_VALUE       = []byte("max value ")
	POINT_COUNT     = []byte("point count ")
	DOC_COUNT       = []byte("doc count ")
	END             = []byte("END")
)

type SimpleTextPointsWriter struct {
	*index.PointsWriterDefault

	dataOut    store.IndexOutput
	scratch    *bytes.Buffer
	writeState *index.SegmentWriteState
	indexFPs   map[string]int64
}

func NewSimpleTextPointsWriter(writeState *index.SegmentWriteState) (*SimpleTextPointsWriter, error) {
	panic("")
}

func (s *SimpleTextPointsWriter) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextPointsWriter) WriteField(fieldInfo *types.FieldInfo, reader index.PointsReader) error {
	//values, err := reader.GetValues(fieldInfo.Name)
	//if err != nil {
	//	return err
	//}
	//
	//config, err := bkd.NewBKDConfig(
	//	fieldInfo.GetPointDimensionCount(),
	//	fieldInfo.GetPointIndexDimensionCount(),
	//	fieldInfo.GetPointNumBytes(),
	//	bkd.DEFAULT_MAX_POINTS_IN_LEAF_NODE,
	//)
	//if err != nil {
	//	return err
	//}

	panic("")
}

func (s *SimpleTextPointsWriter) Finish() error {
	if err := WriteBytes(s.dataOut, END); err != nil {
		return err
	}
	if err := WriteNewline(s.dataOut); err != nil {
		return err
	}
	return WriteChecksum(s.dataOut)
}
