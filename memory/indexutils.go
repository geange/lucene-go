package memory

import (
	"errors"
	"fmt"
	"slices"

	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/tokenattr"
	"github.com/geange/lucene-go/core/util/array"
	"github.com/geange/lucene-go/core/util/bytesutils"
)

func (r *Index) getInfo(fieldName string, fieldType document.IndexableFieldType) (*info, error) {
	if r.frozen {
		return nil, errors.New("cannot call addField() when MemoryIndex is frozen")
	}

	if fieldName == "" {
		return nil, errors.New("fieldName must set")
	}

	var res *info
	v, ok := r.fields.Get(fieldName)
	if ok {
		res = v
	} else {
		res = r.newInfo(r.createFieldInfo(fieldName, r.fields.Size(), fieldType), r.byteBlockPool)
		r.fields.Put(fieldName, res)
	}

	if fieldType.PointDimensionCount() != res.fieldInfo.GetPointDimensionCount() {
		if fieldType.PointDimensionCount() > 0 {
			dimensionCount := fieldType.PointDimensionCount()
			indexDimensionCount := fieldType.PointIndexDimensionCount()
			numBytes := fieldType.PointNumBytes()

			if err := res.fieldInfo.SetPointDimensions(dimensionCount, indexDimensionCount, numBytes); err != nil {
				return nil, err
			}
		}
	}

	if fieldType.DocValuesType() != res.fieldInfo.GetDocValuesType() {
		if fieldType.DocValuesType() != document.DOC_VALUES_TYPE_NONE {
			err := res.fieldInfo.SetDocValuesType(fieldType.DocValuesType())
			if err != nil {
				return nil, err
			}
		}
	}

	return res, nil
}

func (r *Index) storeTerms(info *info, tokenStream analysis.TokenStream, positionIncrementGap, offsetGap int) error {
	pos := -1
	offset := 0
	if info.numTokens > 0 {
		pos = info.lastPosition + positionIncrementGap
		offset = info.lastOffset + offsetGap
	}

	stream := tokenStream

	termAtt := stream.AttributeSource().CharTerm()
	posIncrAttribute := stream.AttributeSource().PositionIncrement()
	offsetAtt := stream.AttributeSource().Offset()
	//packedAttr := stream.AttributeSource().PackedTokenAttribute()
	//bytesAttr := stream.AttributeSource().BytesTerm()
	payloadAtt := stream.AttributeSource().Payload()

	if err := stream.Reset(); err != nil {
		return err
	}

	for {
		ok, err := stream.IncrementToken()
		if err != nil {
			return err
		}
		if !ok {
			break
		}

		info.numTokens++
		posIncr := posIncrAttribute.GetPositionIncrement()
		if posIncr == 0 {
			info.numOverlapTokens++
		}

		pos += posIncr
		ord, err := info.terms.Add([]byte(string(termAtt.Buffer())))
		if err != nil {
			return err
		}
		if ord < 0 {
			ord = (-ord) - 1
			r.postingsWriter.Reset(info.sliceArray.end[ord])
		} else {
			info.sliceArray.start[ord] = r.postingsWriter.StartNewSlice()
		}
		info.sliceArray.freq[ord]++
		info.maxTermFrequency = max(info.maxTermFrequency, info.sliceArray.freq[ord])
		info.sumTotalTermFreq++
		r.postingsWriter.WriteInt(pos)
		if r.storeOffsets {
			r.postingsWriter.WriteInt(offsetAtt.StartOffset() + offset)
			r.postingsWriter.WriteInt(offsetAtt.EndOffset() + offset)
		}

		if r.storePayloads {
			payload := payloadAtt.(tokenattr.PayloadAttr).GetPayload()
			pIndex := 0
			if payload == nil || len(payload) == 0 {
				pIndex = -1
			} else {
				pIndex = r.payloadsBytesRefs.Append(payload)
			}
			r.postingsWriter.WriteInt(pIndex)
		}
		info.sliceArray.end[ord] = r.postingsWriter.GetCurrentOffset()
	}

	if err := stream.End(); err != nil {
		return err
	}

	if info.numTokens > 0 {
		info.lastPosition = pos
		info.lastOffset = offsetAtt.EndOffset() + offset
	}

	return nil
}

func (r *Index) storeDocValues(info *info, docValuesType document.DocValuesType, field document.IndexableField) error {
	fieldName := info.fieldInfo.Name()
	existingDocValuesType := info.fieldInfo.GetDocValuesType()
	if existingDocValuesType == document.DOC_VALUES_TYPE_NONE {
		// first time we add doc values for this field:
		info.fieldInfo = document.NewFieldInfo(
			info.fieldInfo.Name(), info.fieldInfo.Number(), info.fieldInfo.HasVectors(),
			info.fieldInfo.HasPayloads(), info.fieldInfo.HasPayloads(), info.fieldInfo.GetIndexOptions(),
			docValuesType, -1, info.fieldInfo.Attributes(), info.fieldInfo.GetPointDimensionCount(),
			info.fieldInfo.GetPointIndexDimensionCount(), info.fieldInfo.GetPointNumBytes(),
			info.fieldInfo.IsSoftDeletesField())
	} else if existingDocValuesType != docValuesType {
		return fmt.Errorf(`can't add ["%v"] doc values field ["%v"], because ["%v"] doc values field already exists`,
			docValuesType, fieldName, existingDocValuesType,
		)
	}
	switch docValuesType {
	case document.DOC_VALUES_TYPE_NUMERIC:
		value, err := field.I64Value()
		if err != nil {
			return err
		}
		info.numericProducer.dvLongValues = []int{int(value)}
		info.numericProducer.count++
	case document.DOC_VALUES_TYPE_SORTED_NUMERIC:
		if info.numericProducer.dvLongValues == nil {
			info.numericProducer.dvLongValues = make([]int, 4)
		}

		value, err := field.I64Value()
		if err != nil {
			return err
		}
		info.numericProducer.dvLongValues = array.Grow(info.numericProducer.dvLongValues, info.numericProducer.count+10)
		info.numericProducer.dvLongValues[info.numericProducer.count] = int(value)
		info.numericProducer.count++
	case document.DOC_VALUES_TYPE_BINARY:
		if info.binaryProducer.dvBytesValuesSet != nil {
			return fmt.Errorf("only one value per field allowed for [%s] doc values field [%s]", docValuesType, fieldName)
		}
		bytesHash, err := bytesutils.NewBytesHash(r.byteBlockPool)
		if err != nil {
			return err
		}
		info.binaryProducer.dvBytesValuesSet = bytesHash

		value, err := field.BytesValue()
		if err != nil {
			return err
		}

		_, err = info.binaryProducer.dvBytesValuesSet.Add(value)
		if err != nil {
			return err
		}
	case document.DOC_VALUES_TYPE_SORTED:
		if info.binaryProducer.dvBytesValuesSet != nil {
			return fmt.Errorf("only one value per field allowed for [%s] doc values field [%s]", docValuesType, fieldName)
		}
		bytesHash, err := bytesutils.NewBytesHash(r.byteBlockPool)
		if err != nil {
			return err
		}
		info.binaryProducer.dvBytesValuesSet = bytesHash

		value, err := field.BytesValue()
		if err != nil {
			return err
		}

		_, err = info.binaryProducer.dvBytesValuesSet.Add(value)
		if err != nil {
			return err
		}
	case document.DOC_VALUES_TYPE_SORTED_SET:
		if info.binaryProducer.dvBytesValuesSet == nil {
			bytesHash, err := bytesutils.NewBytesHash(r.byteBlockPool)
			if err != nil {
				return err
			}
			info.binaryProducer.dvBytesValuesSet = bytesHash
		}

		value, err := field.BytesValue()
		if err != nil {
			return err
		}

		_, err = info.binaryProducer.dvBytesValuesSet.Add(value)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown doc values type [%s]", docValuesType)
	}

	return nil
}

func (r *Index) createFieldInfo(fieldName string, ord int, fieldType document.IndexableFieldType) *document.FieldInfo {
	indexOptions := document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
	if r.storeOffsets {
		indexOptions = document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	}

	return document.NewFieldInfo(fieldName, ord, fieldType.StoreTermVectors(), fieldType.OmitNorms(), r.storePayloads,
		indexOptions, fieldType.DocValuesType(), -1, map[string]string{},
		fieldType.PointDimensionCount(), fieldType.PointIndexDimensionCount(), fieldType.PointNumBytes(), false)
}

func (r *Index) storePointValues(info *info, pointValue []byte) error {
	if len(info.pointValues) == 0 {
		info.pointValues = make([][]byte, 4)
	}
	info.pointValues = array.Grow(info.pointValues, info.pointValuesCount+1)
	info.pointValues[info.pointValuesCount] = slices.Clone(pointValue)
	info.pointValuesCount++
	return nil
}

func (r *Index) newInfo(docFieldInfo *document.FieldInfo, pool *bytesutils.BlockPool) *info {
	sliceArray := newSliceByteStartArray(bytesutils.DefaultCapacity)

	byteHash, err := bytesutils.NewBytesHash(
		pool,
		bytesutils.WithCapacity(bytesutils.DefaultCapacity),
		bytesutils.WithStartArray(sliceArray),
	)
	if err != nil {
		return nil
	}

	return &info{
		index:           r,
		fieldInfo:       docFieldInfo,
		terms:           byteHash,
		sliceArray:      sliceArray,
		sortedTerms:     make([]int, 0),
		binaryProducer:  newBinaryDocValuesProducer(),
		numericProducer: newNumericDocValuesProducer(),
		//pointValues:     make([][]byte, 0),
		minPackedValue: make([]byte, 0),
		maxPackedValue: make([]byte, 0),
	}
}
