package simpletext

import (
	"bytes"
	"context"
	index2 "github.com/geange/lucene-go/core/interface/index"
	"strconv"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.FieldInfosFormat = &FieldInfosFormat{}

const (
	// FIELD_INFOS_EXTENSION Extension of field infos
	FIELD_INFOS_EXTENSION = "inf"
)

var (
	NUMFIELDS       = []byte("number of fields ")
	NAME            = []byte("  name ")
	NUMBER          = []byte("  number ")
	STORETV         = []byte("  term vectors ")
	STORETVPOS      = []byte("  term vector positions ")
	STORETVOFF      = []byte("  term vector offsets ")
	PAYLOADS        = []byte("  payloads ")
	NORMS           = []byte("  norms ")
	DOCVALUES       = []byte("  doc values ")
	DOCVALUES_GEN   = []byte("  doc values gen ")
	INDEXOPTIONS    = []byte("  index options ")
	NUM_ATTS        = []byte("  attributes ")
	ATT_KEY         = []byte("    key ")
	ATT_VALUE       = []byte("    value ")
	DATA_DIM_COUNT  = []byte("  data dimensional count ")
	INDEX_DIM_COUNT = []byte("  index dimensional count ")
	DIM_NUM_BYTES   = []byte("  dimensional num bytes ")
	SOFT_DELETES    = []byte("  soft-deletes ")
)

// FieldInfosFormat
// plaintext field infos format
// FOR RECREATIONAL USE ONLY
// lucene.experimental
type FieldInfosFormat struct {
}

func NewSimpleTextFieldInfosFormat() *FieldInfosFormat {
	return &FieldInfosFormat{}
}

func (s *FieldInfosFormat) Read(ctx context.Context, directory store.Directory, segmentInfo *index.SegmentInfo, segmentSuffix string, ioContext *store.IOContext) (index2.FieldInfos, error) {
	fileName := store.SegmentFileName(segmentInfo.Name(), segmentSuffix, FIELD_INFOS_EXTENSION)
	input, err := store.OpenChecksumInput(directory, fileName)
	if err != nil {
		return nil, err
	}
	defer func() {
		input.Close()
	}()

	scratch := new(bytes.Buffer)

	r := utils.NewTextReader(input, scratch)

	size, err := r.ParseInt(NUMFIELDS)
	if err != nil {
		return nil, err
	}

	infos := make([]*document.FieldInfo, 0, size)

	for i := 0; i < size; i++ {
		name, err := r.ReadLabel(NAME)
		if err != nil {
			return nil, err
		}

		fieldNumber, err := r.ParseInt(NUMBER)
		if err != nil {
			return nil, err
		}

		value, err := r.ParseString(INDEXOPTIONS)
		if err != nil {
			return nil, err
		}
		indexOptions := document.StringToIndexOptions(value)

		storeTermVector, err := r.ParseBoolPrefix(STORETV)
		if err != nil {
			return nil, err
		}

		storePayloads, err := r.ParseBoolPrefix(PAYLOADS)
		if err != nil {
			return nil, err
		}

		norms, err := r.ParseBoolPrefix(NORMS)
		if err != nil {
			return nil, err
		}
		omitNorms := !norms

		value, err = r.ReadLabel(DOCVALUES)
		if err != nil {
			return nil, err
		}
		docValuesType := document.StringToDocValuesType(value)

		dvGen, err := r.ParseInt(DOCVALUES_GEN)
		if err != nil {
			return nil, err
		}

		numAtts, err := r.ParseInt(NUM_ATTS)
		if err != nil {
			return nil, err
		}

		atts := make(map[string]string, numAtts)

		for j := 0; j < numAtts; j++ {
			attKey, err := r.ReadLabel(ATT_KEY)
			if err != nil {
				return nil, err
			}

			attValue, err := r.ReadLabel(ATT_VALUE)
			if err != nil {
				return nil, err
			}

			atts[attKey] = attValue
		}

		dimensionalCount, err := r.ParseInt(DATA_DIM_COUNT)
		if err != nil {
			return nil, err
		}

		indexDimensionalCount, err := r.ParseInt(INDEX_DIM_COUNT)
		if err != nil {
			return nil, err
		}

		dimensionalNumBytes, err := r.ParseInt(DIM_NUM_BYTES)
		if err != nil {
			return nil, err
		}

		isSoftDeletesField, err := r.ParseBoolPrefix(SOFT_DELETES)
		if err != nil {
			return nil, err
		}

		info := document.NewFieldInfo(name, fieldNumber, storeTermVector,
			omitNorms, storePayloads, indexOptions, docValuesType, int64(dvGen), atts,
			dimensionalCount, indexDimensionalCount, dimensionalNumBytes, isSoftDeletesField)
		infos = append(infos, info)
	}

	if err := utils.CheckFooter(input); err != nil {
		return nil, err
	}

	fieldInfos := index.NewFieldInfos(infos)
	return fieldInfos, nil
}

func (s *FieldInfosFormat) Write(ctx context.Context, directory store.Directory, segmentInfo *index.SegmentInfo, segmentSuffix string, infos index2.FieldInfos, ioContext *store.IOContext) error {
	fileName := store.SegmentFileName(segmentInfo.Name(), segmentSuffix, FIELD_INFOS_EXTENSION)
	out, err := directory.CreateOutput(nil, fileName)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	w := utils.NewTextWriter(out)

	if err := w.WriteLabelInt(NUMFIELDS, infos.Size()); err != nil {
		return err
	}

	for _, fi := range infos.List() {
		if err := w.WriteLabelString(NAME, fi.Name()); err != nil {
			return err
		}

		if err := w.WriteLabelInt(NUMBER, fi.Number()); err != nil {
			return err
		}

		indexOptions := fi.GetIndexOptions()
		//assert indexOptions.compareTo(IndexOptions.DOCS_AND_FREQS_AND_POSITIONS) >= 0 || !fi.hasPayloads();
		if err := w.WriteLabelString(INDEXOPTIONS, indexOptions.String()); err != nil {
			return err
		}

		if err := w.WriteLabelBool(STORETV, fi.HasVectors()); err != nil {
			return err
		}

		if err := w.WriteLabelBool(PAYLOADS, fi.HasPayloads()); err != nil {
			return err
		}

		if err := w.WriteLabelBool(NORMS, !fi.OmitsNorms()); err != nil {
			return err
		}

		if err := w.WriteLabelString(DOCVALUES, fi.GetDocValuesType().String()); err != nil {
			return err
		}

		if err := w.WriteLabelLong(DOCVALUES_GEN, fi.GetDocValuesGen()); err != nil {
			return err
		}

		atts := fi.Attributes()
		numAtts := len(atts)
		if err := w.WriteLabelInt(NUM_ATTS, numAtts); err != nil {
			return err
		}

		if numAtts > 0 {
			for k, v := range atts {
				if err := w.WriteLabelString(ATT_KEY, k); err != nil {
					return err
				}
				if err := w.WriteLabelString(ATT_VALUE, v); err != nil {
					return err
				}
			}
		}

		if err := w.WriteLabelInt(DATA_DIM_COUNT, fi.GetPointDimensionCount()); err != nil {
			return err
		}

		if err := w.WriteLabelInt(INDEX_DIM_COUNT, fi.GetPointIndexDimensionCount()); err != nil {
			return err
		}

		if err := w.WriteLabelInt(DIM_NUM_BYTES, fi.GetPointNumBytes()); err != nil {
			return err
		}

		if err := w.WriteLabelBool(SOFT_DELETES, fi.IsSoftDeletesField()); err != nil {
			return err
		}
	}

	return utils.WriteChecksum(out)
}

func writeValue[T int | int64 | string | bool | []byte](out store.DataOutput, label []byte, value T) error {
	if err := utils.WriteBytes(out, label); err != nil {
		return err
	}

	obj := any(value)

	switch obj.(type) {
	case int:
		if err := utils.WriteString(out, strconv.Itoa(obj.(int))); err != nil {
			return err
		}
	case int32:
		if err := utils.WriteString(out, strconv.Itoa(int(obj.(int32)))); err != nil {
			return err
		}
	case string:
		if err := utils.WriteString(out, obj.(string)); err != nil {
			return err
		}
	case bool:
		if err := utils.WriteString(out, strconv.FormatBool(obj.(bool))); err != nil {
			return err
		}
	case int64:
		if err := utils.WriteString(out, strconv.FormatInt(obj.(int64), 10)); err != nil {
			return err
		}
	case []byte:
		if err := utils.WriteBytes(out, obj.([]byte)); err != nil {
			return err
		}
	}

	return utils.NewLine(out)
}
