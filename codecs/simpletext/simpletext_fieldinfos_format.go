package simpletext

import (
	"bytes"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"strconv"
)

var _ index.FieldInfosFormat = &SimpleTextFieldInfosFormat{}

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

// SimpleTextFieldInfosFormat
// plaintext field infos format
// FOR RECREATIONAL USE ONLY
// lucene.experimental
type SimpleTextFieldInfosFormat struct {
}

func NewSimpleTextFieldInfosFormat() *SimpleTextFieldInfosFormat {
	return &SimpleTextFieldInfosFormat{}
}

func (s *SimpleTextFieldInfosFormat) Read(directory store.Directory, segmentInfo *index.SegmentInfo, segmentSuffix string, ctx *store.IOContext) (*index.FieldInfos, error) {
	fileName := store.SegmentFileName(segmentInfo.Name(), segmentSuffix, FIELD_INFOS_EXTENSION)
	input, err := store.OpenChecksumInput(directory, fileName, ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		input.Close()
	}()
	scratch := new(bytes.Buffer)

	value, err := readValue(input, NUMFIELDS, scratch)
	if err != nil {
		return nil, err
	}
	size, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	infos := make([]*types.FieldInfo, 0, size)

	for i := 0; i < size; i++ {
		value, err := readValue(input, NAME, scratch)
		if err != nil {
			return nil, err
		}
		name := value

		value, err = readValue(input, NUMBER, scratch)
		if err != nil {
			return nil, err
		}
		fieldNumber, _ := strconv.Atoi(value)

		value, err = readValue(input, INDEXOPTIONS, scratch)
		if err != nil {
			return nil, err
		}
		indexOptions := types.StringToIndexOptions(value)

		value, err = readValue(input, STORETV, scratch)
		if err != nil {
			return nil, err
		}
		storeTermVector, _ := strconv.ParseBool(value)

		value, err = readValue(input, PAYLOADS, scratch)
		if err != nil {
			return nil, err
		}
		storePayloads, _ := strconv.ParseBool(value)

		value, err = readValue(input, NORMS, scratch)
		if err != nil {
			return nil, err
		}
		v, _ := strconv.ParseBool(value)
		omitNorms := !v

		value, err = readValue(input, DOCVALUES, scratch)
		if err != nil {
			return nil, err
		}
		docValuesType := types.StringToDocValuesType(value)

		value, err = readValue(input, DOCVALUES_GEN, scratch)
		if err != nil {
			return nil, err
		}
		dvGen, _ := strconv.Atoi(value)

		value, err = readValue(input, NUM_ATTS, scratch)
		if err != nil {
			return nil, err
		}
		numAtts, _ := strconv.Atoi(value)

		atts := make(map[string]string, numAtts)

		for j := 0; j < numAtts; j++ {
			key, err := readValue(input, ATT_KEY, scratch)
			if err != nil {
				return nil, err
			}

			value, err := readValue(input, ATT_VALUE, scratch)
			if err != nil {
				return nil, err
			}

			atts[key] = value
		}

		value, err = readValue(input, DATA_DIM_COUNT, scratch)
		if err != nil {
			return nil, err
		}
		dimensionalCount, _ := strconv.Atoi(value)

		value, err = readValue(input, INDEX_DIM_COUNT, scratch)
		if err != nil {
			return nil, err
		}
		indexDimensionalCount, _ := strconv.Atoi(value)

		value, err = readValue(input, DIM_NUM_BYTES, scratch)
		if err != nil {
			return nil, err
		}
		dimensionalNumBytes, _ := strconv.Atoi(value)

		value, err = readValue(input, SOFT_DELETES, scratch)
		if err != nil {
			return nil, err
		}
		isSoftDeletesField, _ := strconv.ParseBool(value)

		info := types.NewFieldInfo(name, fieldNumber, storeTermVector,
			omitNorms, storePayloads, indexOptions, docValuesType, int64(dvGen), atts,
			dimensionalCount, indexDimensionalCount, dimensionalNumBytes, isSoftDeletesField)
		infos = append(infos, info)
	}

	if err := CheckFooter(input); err != nil {
		return nil, err
	}

	fieldInfos := index.NewFieldInfos(infos)
	return fieldInfos, nil
}

func (s *SimpleTextFieldInfosFormat) Write(directory store.Directory, segmentInfo *index.SegmentInfo,
	segmentSuffix string, infos *index.FieldInfos, context *store.IOContext) error {
	fileName := store.SegmentFileName(segmentInfo.Name(), segmentSuffix, FIELD_INFOS_EXTENSION)
	out, err := directory.CreateOutput(fileName, context)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if err := writeValue(out, NUMFIELDS, strconv.Itoa(infos.Size())); err != nil {
		return err
	}

	for _, fi := range infos.List() {
		if err := writeValue(out, NAME, fi.Name); err != nil {
			return err
		}

		if err := writeValue(out, NUMBER, fi.Number); err != nil {
			return err
		}

		indexOptions := fi.GetIndexOptions()
		//assert indexOptions.compareTo(IndexOptions.DOCS_AND_FREQS_AND_POSITIONS) >= 0 || !fi.hasPayloads();
		if err := writeValue(out, INDEXOPTIONS, indexOptions.String()); err != nil {
			return err
		}

		if err := writeValue(out, STORETV, fi.HasVectors()); err != nil {
			return err
		}

		if err := writeValue(out, PAYLOADS, fi.HasPayloads()); err != nil {
			return err
		}

		if err := writeValue(out, NORMS, !fi.OmitsNorms()); err != nil {
			return err
		}

		if err := writeValue(out, DOCVALUES, fi.GetDocValuesType().String()); err != nil {
			return err
		}

		if err := writeValue(out, DOCVALUES_GEN, fi.GetDocValuesGen()); err != nil {
			return err
		}

		atts := fi.Attributes()
		numAtts := len(atts)
		if err := writeValue(out, NUM_ATTS, numAtts); err != nil {
			return err
		}

		if numAtts > 0 {
			for k, v := range atts {
				if err := writeValue(out, ATT_KEY, k); err != nil {
					return err
				}
				if err := writeValue(out, ATT_VALUE, v); err != nil {
					return err
				}
			}
		}

		if err := writeValue(out, DATA_DIM_COUNT, fi.GetPointDimensionCount()); err != nil {
			return err
		}

		if err := writeValue(out, INDEX_DIM_COUNT, fi.GetPointIndexDimensionCount()); err != nil {
			return err
		}

		if err := writeValue(out, DIM_NUM_BYTES, fi.GetPointNumBytes()); err != nil {
			return err
		}

		if err := writeValue(out, SOFT_DELETES, fi.IsSoftDeletesField()); err != nil {
			return err
		}
	}

	return WriteChecksum(out)

}

func writeValue[T int | int64 | string | bool](out store.DataOutput, label []byte, value T) error {
	if err := WriteBytes(out, label); err != nil {
		return err
	}

	obj := any(value)

	switch obj.(type) {
	case int:
		if err := WriteString(out, strconv.Itoa(obj.(int))); err != nil {
			return err
		}
	case string:
		if err := WriteString(out, obj.(string)); err != nil {
			return err
		}
	case bool:
		if err := WriteString(out, strconv.FormatBool(obj.(bool))); err != nil {
			return err
		}
	case int64:
		if err := WriteString(out, strconv.FormatInt(obj.(int64), 10)); err != nil {
			return err
		}
	}

	return WriteNewline(out)
}
