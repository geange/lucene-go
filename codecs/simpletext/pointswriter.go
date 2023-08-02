package simpletext

import (
	"bytes"
	"context"
	"errors"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/document"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/bkd"
)

var _ index.PointsWriter = &PointsWriter{}

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

type PointsWriter struct {
	*coreIndex.BasePointsWriter

	dataOut    store.IndexOutput
	scratch    *bytes.Buffer
	writeState *index.SegmentWriteState
	indexFPs   map[string]int64
}

func NewSimpleTextPointsWriter(ctx context.Context, writeState *index.SegmentWriteState) (*PointsWriter, error) {
	fileName := store.SegmentFileName(writeState.SegmentInfo.Name(), writeState.SegmentSuffix, POINT_EXTENSION)
	out, err := writeState.Directory.CreateOutput(ctx, fileName)
	if err != nil {
		return nil, err
	}
	writer := &PointsWriter{
		BasePointsWriter: nil,
		dataOut:          out,
		scratch:          new(bytes.Buffer),
		writeState:       writeState,
		indexFPs:         make(map[string]int64),
	}
	writer.BasePointsWriter = &coreIndex.BasePointsWriter{
		WriteField: writer.WriteField,
		Finish:     writer.Finish,
	}
	return writer, nil
}

func (s *PointsWriter) Close() error {
	if s.dataOut == nil {
		return nil
	}

	if err := s.dataOut.Close(); err != nil {
		return err
	}
	s.dataOut = nil

	fileName := store.SegmentFileName(s.writeState.SegmentInfo.Name(),
		s.writeState.SegmentSuffix, POINT_INDEX_EXTENSION)

	indexOut, err := s.writeState.Directory.CreateOutput(nil, fileName)
	if err != nil {
		return err
	}
	count := len(s.indexFPs)

	w := utils.NewTextWriter(indexOut)
	if err := w.Bytes(FIELD_COUNT); err != nil {
		return err
	}
	if err := w.Int(count); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	for k, v := range s.indexFPs {
		if err := w.Bytes(FIELD_FP_NAME); err != nil {
			return err
		}
		if err := w.String(k); err != nil {
			return err
		}
		if err := w.NewLine(); err != nil {
			return err
		}

		if err := w.Bytes(FIELD_FP); err != nil {
			return err
		}
		if err := w.Int(int(v)); err != nil {
			return err
		}
		if err := w.NewLine(); err != nil {
			return err
		}
	}

	if err := w.Checksum(); err != nil {
		return err
	}
	return indexOut.Close()
}

func (s *PointsWriter) WriteField(ctx context.Context, fieldInfo *document.FieldInfo, reader index.PointsReader) error {
	values, err := reader.GetValues(nil, fieldInfo.Name())
	if err != nil {
		return err
	}

	config, err := bkd.NewConfig(
		fieldInfo.GetPointDimensionCount(),
		fieldInfo.GetPointIndexDimensionCount(),
		fieldInfo.GetPointNumBytes(),
		bkd.DEFAULT_MAX_POINTS_IN_LEAF_NODE,
	)
	if err != nil {
		return err
	}

	maxDoc, err := s.writeState.SegmentInfo.MaxDoc()
	if err != nil {
		return err
	}
	writer := NewBKDWriter(maxDoc,
		s.writeState.Directory,
		s.writeState.SegmentInfo.Name(),
		config,
		DEFAULT_MAX_MB_SORT_IN_HEAP,
		values.Size())

	if err := values.Intersect(ctx, &types.BytesVisitor{
		VisitFn: func(docID int) error {
			return errors.New("illegal State")
		},
		VisitLeafFn: func(docID int, packedValue []byte) error {
			return writer.Add(packedValue, docID)
		},
		CompareFn: func(minPackedValue, maxPackedValue []byte) types.Relation {
			return types.CELL_CROSSES_QUERY
		},
		GrowFn: func(count int) {
		},
	}); err != nil {
		return err
	}

	// We could have 0 points on merge since all docs with points may be deleted:
	if writer.GetPointCount() > 0 {
		fp, err := writer.Finish(s.dataOut)
		if err != nil {
			return err
		}
		s.indexFPs[fieldInfo.Name()] = fp
	}

	return nil
}

func (s *PointsWriter) Finish() error {
	if err := utils.WriteBytes(s.dataOut, END); err != nil {
		return err
	}
	if err := utils.NewLine(s.dataOut); err != nil {
		return err
	}
	return utils.WriteChecksum(s.dataOut)
}
