package lucene60

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/document"
	index2 "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.FieldInfosFormat = &FieldInfosFormat{}

type FieldInfosFormat struct {
}

func NewFieldInfosFormat() *FieldInfosFormat {
	return &FieldInfosFormat{}
}

func (f *FieldInfosFormat) Read(ctx context.Context, directory store.Directory, segmentInfo index.SegmentInfo, segmentSuffix string, ioContext *store.IOContext) (index.FieldInfos, error) {
	fileName := store.SegmentFileName(segmentInfo.Name(), segmentSuffix, FNM_EXTENSION)
	input, err := store.OpenChecksumInput(ctx, directory, fileName)
	if err != nil {
		return nil, err
	}

	version, err := codecs.CheckIndexHeader(ctx, input,
		FNM_CODEC_NAME, FNM_FORMAT_START, FNM_FORMAT_CURRENT,
		segmentInfo.GetID(), segmentSuffix)
	if err != nil {
		return nil, err
	}

	size, err := input.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	infos := make([]*document.FieldInfo, size)

	// previous field's attribute map, we share when possible:
	lastAttributes := make(map[string]string)

	for i := 0; i < int(size); i++ {
		name, err := input.ReadString(ctx)
		if err != nil {
			return nil, err
		}
		fieldNumber, err := input.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		if fieldNumber < 0 {
			return nil, fmt.Errorf("invalid field number for field: %s, fieldNumber=%d", name, fieldNumber)
		}

		bits, err := input.ReadByte()
		if err != nil {
			return nil, err
		}
		storeTermVector := (bits & byte(FNM_STORE_TERMVECTOR)) != 0
		omitNorms := (bits & byte(FNM_OMIT_NORMS)) != 0
		storePayloads := (bits & byte(FNM_STORE_PAYLOADS)) != 0
		isSoftDeletesField := (bits & byte(FNM_SOFT_DELETES_FIELD)) != 0

		byteIndexOptions, err := input.ReadByte()
		if err != nil {
			return nil, err
		}
		indexOptions, err := getIndexOptions(byteIndexOptions)
		if err != nil {
			return nil, err
		}

		byteDocValuesType, err := input.ReadByte()
		if err != nil {
			return nil, err
		}
		docValuesType, err := getDocValuesType(byteDocValuesType)
		if err != nil {
			return nil, err
		}

		dvGen, err := input.ReadUint64(ctx)
		if err != nil {
			return nil, err
		}
		attributes, err := input.ReadMapOfStrings(ctx)
		if err != nil {
			return nil, err
		}
		// just use the last field's map if its the same
		if reflect.DeepEqual(attributes, lastAttributes) {
			attributes = lastAttributes
		}
		lastAttributes = attributes

		pointDataDimensionCount, err := input.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}

		pointIndexDimensionCount := pointDataDimensionCount
		var pointNumBytes uint64
		if pointDataDimensionCount != 0 {
			if version >= FNM_FORMAT_SELECTIVE_INDEXING {
				count, err := input.ReadUvarint(ctx)
				if err != nil {
					return nil, err
				}
				pointIndexDimensionCount = count
			}
			bs, err := input.ReadUvarint(ctx)
			if err != nil {
				return nil, err
			}
			pointNumBytes = bs
		} else {
			pointNumBytes = 0
		}

		infos[i] = document.NewFieldInfo(name, int(fieldNumber), storeTermVector, omitNorms, storePayloads,
			indexOptions, docValuesType, int64(dvGen), attributes,
			int(pointDataDimensionCount), int(pointIndexDimensionCount), int(pointNumBytes), isSoftDeletesField)
	}
	if _, err := codecs.CheckFooter(ctx, input); err != nil {
		return nil, err
	}
	return index2.NewFieldInfos(infos), nil
}

func (f *FieldInfosFormat) Write(ctx context.Context, directory store.Directory, segmentInfo index.SegmentInfo, segmentSuffix string, infos index.FieldInfos, ioContext *store.IOContext) error {
	fileName := store.SegmentFileName(segmentInfo.Name(), segmentSuffix, FNM_EXTENSION)
	output, err := directory.CreateOutput(ctx, fileName)
	if err != nil {
		return err
	}
	if err := codecs.WriteIndexHeader(ctx, output, FNM_CODEC_NAME, FNM_FORMAT_CURRENT,
		segmentInfo.GetID(), segmentSuffix); err != nil {
		return err
	}

	if err := output.WriteUvarint(ctx, uint64(infos.Size())); err != nil {
		return err
	}

	for _, fi := range infos.List() {
		if err := fi.CheckConsistency(); err != nil {
			return err
		}

		if err := output.WriteString(ctx, fi.Name()); err != nil {
			return err
		}
		if err := output.WriteUvarint(ctx, uint64(fi.Number())); err != nil {
			return err
		}

		bits := byte(0x0)
		if fi.HasVectors() {
			bits |= byte(FNM_STORE_TERMVECTOR)
		}
		if fi.OmitsNorms() {
			bits |= byte(FNM_OMIT_NORMS)
		}
		if fi.HasPayloads() {
			bits |= byte(FNM_STORE_PAYLOADS)
		}
		if fi.IsSoftDeletesField() {
			bits |= byte(FNM_SOFT_DELETES_FIELD)
		}
		if err := output.WriteByte(bits); err != nil {
			return err
		}

		if err := output.WriteByte(indexOptionsByte(fi.GetIndexOptions())); err != nil {
			return err
		}

		// pack the DV type and hasNorms in one byte
		if err := output.WriteByte(docValuesByte(fi.GetDocValuesType())); err != nil {
			return err
		}
		if err := output.WriteUint64(ctx, uint64(fi.GetDocValuesGen())); err != nil {
			return err
		}
		if err := output.WriteMapOfStrings(ctx, fi.Attributes()); err != nil {
			return err
		}
		if err := output.WriteUvarint(ctx, uint64(fi.GetPointDimensionCount())); err != nil {
			return err
		}
		if fi.GetPointDimensionCount() != 0 {
			if err := output.WriteUvarint(ctx, uint64(fi.GetPointIndexDimensionCount())); err != nil {
				return err
			}
			if err := output.WriteUvarint(ctx, uint64(fi.GetPointNumBytes())); err != nil {
				return err
			}
		}
	}
	return codecs.WriteFooter(ctx, output)
}

const (
	FNM_EXTENSION = "fnm" // Extension of field infos

	// Codec header
	FNM_CODEC_NAME                = "Lucene60FieldInfos"
	FNM_FORMAT_START              = 0
	FNM_FORMAT_SOFT_DELETES       = 1
	FNM_FORMAT_SELECTIVE_INDEXING = 2
	FNM_FORMAT_CURRENT            = FNM_FORMAT_SELECTIVE_INDEXING

	// Field flags
	FNM_STORE_TERMVECTOR   = 0x1
	FNM_OMIT_NORMS         = 0x2
	FNM_STORE_PAYLOADS     = 0x4
	FNM_SOFT_DELETES_FIELD = 0x8
)

func docValuesByte(dType document.DocValuesType) byte {
	return byte(dType)
}

func getDocValuesType(b byte) (document.DocValuesType, error) {
	dType := document.DocValuesType(b)
	switch dType {
	case document.DOC_VALUES_TYPE_NONE:
	case document.DOC_VALUES_TYPE_NUMERIC:
	case document.DOC_VALUES_TYPE_BINARY:
	case document.DOC_VALUES_TYPE_SORTED:
	case document.DOC_VALUES_TYPE_SORTED_SET:
	case document.DOC_VALUES_TYPE_SORTED_NUMERIC:
	default:
		return 0, errors.New("invalid docvalues byte")
	}
	return dType, nil
}

func indexOptionsByte(indexOptions document.IndexOptions) byte {
	return byte(indexOptions)
}

func getIndexOptions(b byte) (document.IndexOptions, error) {
	opts := document.IndexOptions(b)
	switch opts {
	case document.INDEX_OPTIONS_NONE:
	case document.INDEX_OPTIONS_DOCS:
	case document.INDEX_OPTIONS_DOCS_AND_FREQS:
	case document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS:
	case document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS:
	default:
		return 0, errors.New("invalid IndexOptions byte")
	}
	return opts, nil
}
