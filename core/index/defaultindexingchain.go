package index

import (
	"context"
	"errors"
	"fmt"

	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/bytesref"
	"github.com/geange/lucene-go/core/util/ints"
)

var _ DocConsumer = &DefaultIndexingChain{}

// DefaultIndexingChain
// Default general purpose indexing chain, which handles indexing all types of fields.
type DefaultIndexingChain struct {
	fieldInfos           *FieldInfosBuilder
	termsHash            TermsHash
	docValuesBytePool    *bytesref.BlockPool
	storedFieldsConsumer *StoredFieldsConsumer
	termVectorsWriter    *TermVectorsConsumer

	// 使用map简化理解
	fieldHash map[string]*PerField
	hashMask  int

	totalFieldCount int
	nextFieldGen    int64

	fields []*PerField

	byteBlockAllocator bytesref.Allocator

	indexWriterConfig *liveIndexWriterConfig

	indexCreatedVersionMajor int

	hasHitAbortingException bool
}

func NewDefaultIndexingChain(indexCreatedVersionMajor int, segmentInfo *SegmentInfo, dir store.Directory, fieldInfos *FieldInfosBuilder, indexWriterConfig *liveIndexWriterConfig) *DefaultIndexingChain {

	byteBlockAllocator := newByteBlockAllocator()
	intBlockAllocator := newIntBlockAllocator()

	var storedFieldsConsumer *StoredFieldsConsumer
	var termVectorsWriter *TermVectorsConsumer
	if segmentInfo.GetIndexSort() == nil {
		storedFieldsConsumer = NewStoredFieldsConsumer(indexWriterConfig.GetCodec(), dir, segmentInfo)
		termVectorsWriter = NewTermVectorsConsumer(intBlockAllocator, byteBlockAllocator, dir, segmentInfo, indexWriterConfig.GetCodec())
	}

	indexChain := &DefaultIndexingChain{
		fieldInfos:               fieldInfos,
		termsHash:                NewFreqProxTermsWriter(intBlockAllocator, byteBlockAllocator, termVectorsWriter),
		docValuesBytePool:        bytesref.NewBlockPool(byteBlockAllocator),
		storedFieldsConsumer:     storedFieldsConsumer,
		termVectorsWriter:        termVectorsWriter,
		fieldHash:                make(map[string]*PerField),
		hashMask:                 1,
		totalFieldCount:          0,
		nextFieldGen:             0,
		fields:                   make([]*PerField, 0),
		byteBlockAllocator:       byteBlockAllocator,
		indexWriterConfig:        indexWriterConfig,
		indexCreatedVersionMajor: indexCreatedVersionMajor,
		hasHitAbortingException:  false,
	}

	return indexChain
}

func (d *DefaultIndexingChain) getDocValuesLeafReader() LeafReader {
	reader := &innerLeafReader{
		DocValuesLeafReader: NewDocValuesLeafReader(),
		chain:               d,
	}
	return reader
}

var _ LeafReader = &innerLeafReader{}

type innerLeafReader struct {
	*DocValuesLeafReader
	chain *DefaultIndexingChain
}

func (i *innerLeafReader) GetNumericDocValues(field string) (NumericDocValues, error) {
	pf := i.chain.getPerField(field)
	if pf == nil {
		return nil, nil
	}
	if pf.fieldInfo.GetDocValuesType() == document.DOC_VALUES_TYPE_NUMERIC {
		return pf.docValuesWriter.GetDocValues().(NumericDocValues), nil
	}
	return nil, nil
}

func (i *innerLeafReader) GetBinaryDocValues(field string) (BinaryDocValues, error) {
	pf := i.chain.getPerField(field)
	if pf == nil {
		return nil, nil
	}
	if pf.fieldInfo.GetDocValuesType() == document.DOC_VALUES_TYPE_BINARY {
		return pf.docValuesWriter.GetDocValues().(BinaryDocValues), nil
	}
	return nil, nil
}

func (i *innerLeafReader) GetSortedDocValues(field string) (SortedDocValues, error) {
	pf := i.chain.getPerField(field)
	if pf == nil {
		return nil, nil
	}
	if pf.fieldInfo.GetDocValuesType() == document.DOC_VALUES_TYPE_SORTED {
		return pf.docValuesWriter.GetDocValues().(SortedDocValues), nil
	}
	return nil, nil
}

func (i *innerLeafReader) GetSortedNumericDocValues(field string) (SortedNumericDocValues, error) {
	pf := i.chain.getPerField(field)
	if pf == nil {
		return nil, nil
	}
	if pf.fieldInfo.GetDocValuesType() == document.DOC_VALUES_TYPE_SORTED_NUMERIC {
		return pf.docValuesWriter.GetDocValues().(SortedNumericDocValues), nil
	}
	return nil, nil
}

func (i *innerLeafReader) GetSortedSetDocValues(field string) (SortedSetDocValues, error) {
	pf := i.chain.getPerField(field)
	if pf == nil {
		return nil, nil
	}
	if pf.fieldInfo.GetDocValuesType() == document.DOC_VALUES_TYPE_SORTED_SET {
		return pf.docValuesWriter.GetDocValues().(SortedSetDocValues), nil
	}
	return nil, nil
}

func (i *innerLeafReader) GetFieldInfos() *FieldInfos {
	return i.chain.fieldInfos.Finish()
}

func (d *DefaultIndexingChain) maybeSortSegment(state *SegmentWriteState) (*DocMap, error) {
	indexSort := state.SegmentInfo.GetIndexSort()
	if indexSort == nil {
		return nil, nil
	}
	docValuesReader := d.getDocValuesLeafReader()
	comparators := make([]DocComparator, 0)

	for _, sortField := range indexSort.GetSort() {
		sorter := sortField.GetIndexSorter()
		if sorter == nil {
			return nil, errors.New("cannot sort index using sort field")
		}

		maxDoc, err := state.SegmentInfo.MaxDoc()
		if err != nil {
			return nil, err
		}
		comparator, err := sorter.GetDocComparator(docValuesReader, maxDoc)
		if err != nil {
			return nil, err
		}
		comparators = append(comparators, comparator)
	}

	maxDoc, err := state.SegmentInfo.MaxDoc()
	if err != nil {
		return nil, err
	}
	return SortByComparators(maxDoc, comparators)
}

func (d *DefaultIndexingChain) Flush(ctx context.Context, state *SegmentWriteState) (*DocMap, error) {
	// NOTE: caller (DocumentsWriterPerThread) handles
	// aborting on any exception from this method
	sortMap, err := d.maybeSortSegment(state)
	if err != nil {
		return nil, err
	}
	maxDoc, err := state.SegmentInfo.MaxDoc()
	if err != nil {
		return nil, err
	}

	if err := d.writeNorms(state, sortMap); err != nil {
		return nil, err
	}

	if err := d.writeDocValues(state, sortMap); err != nil {
		return nil, err
	}
	if err := d.writePoints(state, sortMap); err != nil {
		return nil, err
	}

	if err := d.storedFieldsConsumer.Finish(maxDoc); err != nil {
		return nil, err
	}
	if err := d.storedFieldsConsumer.Flush(state, sortMap); err != nil {
		return nil, err
	}

	fieldsToFlush := make(map[string]TermsHashPerField)

	for _, perField := range d.fieldHash {
		fieldsToFlush[perField.fieldInfo.Name()] = perField.termsHashPerField
	}

	readState := NewSegmentReadState(state.Directory, state.SegmentInfo, state.FieldInfos, nil, state.SegmentSuffix)

	var norms NormsProducer
	if readState.FieldInfos.HasNorms() {
		norms, err = state.SegmentInfo.GetCodec().NormsFormat().NormsProducer(readState)
		if err != nil {
			return nil, err
		}
		normsMergeInstance := norms.GetMergeInstance()
		d.termsHash.Flush(fieldsToFlush, state, sortMap, normsMergeInstance)
	}

	err = d.indexWriterConfig.GetCodec().FieldInfosFormat().
		Write(state.Directory, state.SegmentInfo, "", state.FieldInfos, nil)
	if err != nil {
		return nil, err
	}
	return sortMap, nil
}

// Writes all buffered points.
func (d *DefaultIndexingChain) writePoints(state *SegmentWriteState, sortMap *DocMap) error {
	var pointsWriter PointsWriter
	var err error

	for _, perField := range d.fieldHash {
		if perField.pointValuesWriter != nil {
			if perField.fieldInfo.GetPointDimensionCount() == 0 {
				panic("BUG")
				// BUG
			}
			if pointsWriter == nil {
				// lazy init
				format := state.SegmentInfo.GetCodec().PointsFormat()
				if format == nil {
					return errors.New("pointsFormat not found")
				}
				pointsWriter, err = format.FieldsWriter(state)
				if err != nil {
					return err
				}
			}

			if err := perField.pointValuesWriter.Flush(state, sortMap, pointsWriter); err != nil {
				return err
			}
			perField.pointValuesWriter = nil
		}
	}

	if pointsWriter != nil {
		if err := pointsWriter.Finish(); err != nil {
			return err
		}
		if err := pointsWriter.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Writes all buffered doc values (called from Flush).
func (d *DefaultIndexingChain) writeDocValues(state *SegmentWriteState, sortMap *DocMap) error {
	var dvConsumer DocValuesConsumer
	var err error
	for _, perField := range d.fieldHash {
		if perField.docValuesWriter != nil {
			if perField.fieldInfo.GetDocValuesType() == document.DOC_VALUES_TYPE_NONE {
				panic("BUG")
				// BUG
				//throw new AssertionError("segment=" + state.segmentInfo + ": field=\"" + perField.fieldInfo.name + "\" has no docValues but wrote them");
			}
			if dvConsumer == nil {
				// lazy init
				format := state.SegmentInfo.GetCodec().DocValuesFormat()
				dvConsumer, err = format.FieldsConsumer(state)
				if err != nil {
					return err
				}
			}
			if err := perField.docValuesWriter.Flush(state, *sortMap, dvConsumer); err != nil {
				return err
			}
			perField.docValuesWriter = nil
		} else if perField.fieldInfo.GetDocValuesType() != document.DOC_VALUES_TYPE_NONE {
			panic("BUG")
		}
	}

	if dvConsumer != nil {
		if err := dvConsumer.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (d *DefaultIndexingChain) writeNorms(state *SegmentWriteState, sortMap *DocMap) error {
	var normsConsumer NormsConsumer
	var err error
	if state.FieldInfos.HasNorms() {
		normsFormat := state.SegmentInfo.GetCodec().NormsFormat()
		normsConsumer, err = normsFormat.NormsConsumer(state)
		if err != nil {
			return err
		}

		for _, fi := range state.FieldInfos.List() {
			perField := d.getPerField(fi.Name())

			// we must check the final item of omitNorms for the fieldinfo: it could have
			// Changed for this field since the first time we added it.
			if fi.OmitsNorms() == false && fi.GetIndexOptions() != document.INDEX_OPTIONS_NONE {
				maxDoc, err := state.SegmentInfo.MaxDoc()
				if err != nil {
					return err
				}
				perField.norms.Finish(maxDoc)
				if err := perField.norms.Flush(state, sortMap, normsConsumer); err != nil {
					return err
				}
			}
		}
	}

	if normsConsumer != nil {
		if err := normsConsumer.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (d *DefaultIndexingChain) ProcessDocument(ctx context.Context, docId int, doc *document.Document) error {
	// How many indexed field names we've seen (collapses
	// multiple field instances by the same name):
	fieldCount := 0
	fieldGen := d.nextFieldGen
	d.nextFieldGen++

	// NOTE: we need two passes here, in case there are
	// multi-valued fields, because we must process all
	// instances of a given field at once, since the
	// analyzer is free to reuse TokenStream across fields
	// (i.e., we cannot have more than one TokenStream
	// running "at once"):

	if err := d.termsHash.StartDocument(); err != nil {
		return err
	}
	if err := d.startStoredFields(docId); err != nil {
		return err
	}

	var err error

	for _, field := range doc.Fields() {
		fieldCount, err = d.processField(docId, field, fieldGen, fieldCount)
		if err != nil {
			return err
		}

		for i := 0; i < fieldCount; i++ {
			if err = d.fields[i].Finish(docId); err != nil {
				return err
			}
		}
		if err = d.finishStoredFields(); err != nil {
			return err
		}
	}

	return d.termsHash.FinishDocument(docId)
}

func (d *DefaultIndexingChain) processField(docId int,
	field document.IndexableField, fieldGen int64, fieldCount int) (int, error) {

	fieldName := field.Name()
	fieldType := field.FieldType()

	if fieldType.IndexOptions() == -1 {
		return 0, errors.New("indexOptions must not be null")
	}

	var err error
	var fp *PerField

	// Invert indexed fields:
	if fieldType.IndexOptions() != document.INDEX_OPTIONS_NONE {
		fp, err = d.getOrAddField(fieldName, fieldType, true)
		if err != nil {
			return 0, err
		}
		first := fp.fieldGen != fieldGen
		if err := fp.invert(docId, field, first); err != nil {
			return 0, err
		}

		if first {
			d.fields = append(d.fields, fp)
			fieldCount++
			fp.fieldGen = fieldGen
		}
	} else {
		if err := verifyUnIndexedFieldType(fieldName, fieldType); err != nil {
			return 0, err
		}
	}

	// Add stored fields:
	if fieldType.Stored() {
		if fp == nil {
			fp, err = d.getOrAddField(fieldName, fieldType, false)
			if err != nil {
				return 0, err
			}
		}

		if fieldType.Stored() {
			//str, err := document.Str(field.Get())
			//if err == nil {
			//	if len(str) > MAX_STORED_STRING_LENGTH {
			//		return 0, errors.New("stored field too large")
			//	}
			//}

			if err := d.storedFieldsConsumer.writeField(fp.fieldInfo, field); err != nil {
				return 0, err
			}
		}
	}

	dvType := fieldType.DocValuesType()
	if dvType == -1 {
		return 0, errors.New("docValuesType must not be null")
	}

	if dvType != document.DOC_VALUES_TYPE_NONE {
		if fp == nil {
			fp, err = d.getOrAddField(fieldName, fieldType, false)
			if err != nil {
				return 0, err
			}
		}
		if err := d.indexDocValue(docId, fp, dvType, field); err != nil {
			return 0, err
		}
	}

	if fieldType.PointDimensionCount() != 0 {
		if fp == nil {
			fp, err = d.getOrAddField(fieldName, fieldType, false)
			if err != nil {
				return 0, err
			}
		}
		if err := d.indexPoint(docId, fp, field); err != nil {
			return 0, err
		}
	}
	return fieldCount, nil
}

func verifyUnIndexedFieldType(name string, ft document.IndexableFieldType) error {
	// TODO
	return nil
}

// Called from processDocument to index one field's point
func (d *DefaultIndexingChain) indexPoint(docID int, fp *PerField, field document.IndexableField) error {
	pointDimensionCount := field.FieldType().PointDimensionCount()
	pointIndexDimensionCount := field.FieldType().PointIndexDimensionCount()
	dimensionNumBytes := field.FieldType().PointNumBytes()

	// Record dimensions for this field; this setter will throw IllegalArgExc if
	// the dimensions were already set to something different:
	if fp.fieldInfo.GetPointDimensionCount() == 0 {
		d.fieldInfos.globalFieldNumbers.SetDimensions(fp.fieldInfo.Number(), fp.fieldInfo.Name(), pointDimensionCount, pointIndexDimensionCount, dimensionNumBytes)
	}

	if err := fp.fieldInfo.SetPointDimensions(pointDimensionCount, pointIndexDimensionCount, dimensionNumBytes); err != nil {
		return err
	}
	if fp.pointValuesWriter == nil {
		fp.pointValuesWriter = NewPointValuesWriter(fp.fieldInfo)
	}
	bs, err := document.Bytes(field.Get())
	if err != nil {
		return err
	}
	return fp.pointValuesWriter.AddPackedValue(docID, bs)
}

func (d *DefaultIndexingChain) validateIndexSortDVType(indexSort *Sort, fieldToValidate string, dvType document.DocValuesType) error {
	// TODO: fix it

	return nil
}

// Called from processDocument to index one field's doc item
func (d *DefaultIndexingChain) indexDocValue(docID int,
	fp *PerField, dvType document.DocValuesType, field document.IndexableField) error {

	if fp.fieldInfo.GetDocValuesType() == document.DOC_VALUES_TYPE_NONE {
		// This is the first time we are seeing this field indexed with doc values, so we
		// now record the DV type so that any future attempt to (illegally) change
		// the DV type of this field, will throw an IllegalArgExc:
		if d.indexWriterConfig.GetIndexSort() != nil {
			indexSort := d.indexWriterConfig.GetIndexSort()
			if err := d.validateIndexSortDVType(indexSort, fp.fieldInfo.Name(), dvType); err != nil {
				return err
			}
		}

		if err := d.fieldInfos.globalFieldNumbers.
			setDocValuesType(fp.fieldInfo.Number(), fp.fieldInfo.Name(), dvType); err != nil {
			return err
		}
	}

	if err := fp.fieldInfo.SetDocValuesType(dvType); err != nil {
		return err
	}

	switch dvType {
	case document.DOC_VALUES_TYPE_NUMERIC:
		if fp.docValuesWriter == nil {
			fp.docValuesWriter = NewNumericDocValuesWriter(fp.fieldInfo)
		}

		// TODO: 需要设计如何返回具体的数据类型
		obj, ok := field.Number()
		if !ok {
			return errors.New("field value is not number")
		}
		num, err := document.Int64(obj)
		if err != nil {
			return err
		}

		if err := fp.docValuesWriter.(*NumericDocValuesWriter).AddValue(docID, num); err != nil {
			return err
		}

	case document.DOC_VALUES_TYPE_BINARY:
		if fp.docValuesWriter != nil {
			fp.docValuesWriter = NewBinaryDocValuesWriter(fp.fieldInfo)
		}

		bs, err := document.Bytes(field.Get())
		if err != nil {
			return err
		}

		if err := fp.docValuesWriter.(*BinaryDocValuesWriter).AddValue(docID, bs); err != nil {
			return err
		}

	case document.DOC_VALUES_TYPE_SORTED:
		return errors.New("unsupported DocValues.Type")
	case document.DOC_VALUES_TYPE_SORTED_NUMERIC:
		return errors.New("unsupported DocValues.Type")
	case document.DOC_VALUES_TYPE_SORTED_SET:
		return errors.New("unsupported DocValues.Type")
	default:
		return errors.New("unsupported DocValues.Type")
	}
	return errors.New("unrecognized DocValues.Type")
}

// Returns a previously created DefaultIndexingChain.PerField, absorbing the type information from FieldType,
// and creates a new DefaultIndexingChain.PerField if this field name wasn't seen yet.
func (d *DefaultIndexingChain) getOrAddField(name string, fieldType document.IndexableFieldType, invert bool) (*PerField, error) {
	// Make sure we have a PerField allocated
	fp, ok := d.fieldHash[name]
	if !ok {
		fi, err := d.fieldInfos.GetOrAdd(name)
		if err != nil {
			return nil, err
		}

		if err := d.initIndexOptions(fi, fieldType.IndexOptions()); err != nil {
			return nil, err
		}

		for k, v := range fieldType.GetAttributes() {
			fi.PutAttribute(k, v)
		}

		similarity := d.indexWriterConfig.GetSimilarity()
		analyzer := d.indexWriterConfig.GetAnalyzer()
		fp, err = d.NewPerField(d.indexCreatedVersionMajor, fi, invert, similarity, analyzer)
		if err != nil {
			return nil, err
		}

		d.fieldHash[name] = fp
		d.totalFieldCount++
		return fp, nil
	}

	if invert && fp.invertState == nil {
		if err := d.initIndexOptions(fp.fieldInfo, fieldType.IndexOptions()); err != nil {
			return nil, err
		}
		if err := fp.setInvertState(); err != nil {
			return nil, err
		}
	}
	return fp, nil
}

// Returns a previously created DefaultIndexingChain.PerField, or null if this field name wasn't seen yet.
func (d *DefaultIndexingChain) getPerField(name string) *PerField {
	return d.fieldHash[name]
}

func (d *DefaultIndexingChain) initIndexOptions(info *document.FieldInfo, indexOptions document.IndexOptions) error {
	// Messy: must set this here because e.g. FreqProxTermsWriterPerField looks at the initial
	// IndexOptions to decide what arrays it must create).

	// TODO: assert info.getIndexOptions() == IndexOptions.NONE;

	// This is the first time we are seeing this field indexed, so we now
	// record the index options so that any future attempt to (illegally)
	// change the index options of this field, will throw an IllegalArgExc:
	err := d.fieldInfos.globalFieldNumbers.setIndexOptions(info.Number(), info.Name(), indexOptions)
	if err != nil {
		return err
	}
	return info.SetIndexOptions(indexOptions)
}

func (d *DefaultIndexingChain) Abort() error {
	return nil
}

func (d *DefaultIndexingChain) GetHasDocValues(field string) types.DocIdSetIterator {
	perField := d.getPerField(field)
	if perField != nil {
		if perField.fieldInfo.GetDocValuesType() == document.DOC_VALUES_TYPE_NONE {
			return nil
		}
		return perField.docValuesWriter.GetDocValues()
	}
	return nil
}

func (d *DefaultIndexingChain) startStoredFields(docID int) error {
	return d.storedFieldsConsumer.StartDocument(docID)
}

func (d *DefaultIndexingChain) finishStoredFields() error {
	return d.storedFieldsConsumer.FinishDocument()
}

// PerField
// NOTE: not static: accesses at least docState, termsHash.
type PerField struct {
	chain                    *DefaultIndexingChain
	indexCreatedVersionMajor int
	fieldInfo                *document.FieldInfo
	similarity               Similarity
	invertState              *FieldInvertState
	termsHashPerField        TermsHashPerField
	docValuesWriter          DocValuesWriter
	pointValuesWriter        *PointValuesWriter

	// We use this to know when a PerField is seen for the first time in the current document.
	fieldGen int64

	// Used by the hash table
	//next *PerField

	norms       *NormValuesWriter
	tokenStream analysis.TokenStream
	analyzer    analysis.Analyzer
}

func (d *DefaultIndexingChain) NewPerField(indexCreatedVersionMajor int, fieldInfo *document.FieldInfo,
	invert bool, similarity Similarity, analyzer analysis.Analyzer) (*PerField, error) {

	perField := &PerField{
		chain:                    d,
		indexCreatedVersionMajor: indexCreatedVersionMajor,
		fieldInfo:                fieldInfo,
		similarity:               similarity,
		analyzer:                 analyzer,
	}

	if invert {
		if err := perField.setInvertState(); err != nil {
			return nil, err
		}
	}
	return perField, nil
}

func (p *PerField) setInvertState() error {
	p.invertState = NewFieldInvertStateV1(
		p.indexCreatedVersionMajor, p.fieldInfo.Name(), p.fieldInfo.GetIndexOptions())

	var err error
	p.termsHashPerField, err = p.chain.termsHash.AddField(p.invertState, p.fieldInfo)
	if err != nil {
		return err
	}
	if p.fieldInfo.OmitsNorms() == false {
		p.norms = NewNormValuesWriter(p.fieldInfo)
	}
	return nil
}

func (p *PerField) invert(docID int, field document.IndexableField, first bool) error {
	if first {
		p.invertState.Reset()
	}

	fieldType := field.FieldType()
	indexOptions := fieldType.IndexOptions()
	if err := p.fieldInfo.SetIndexOptions(indexOptions); err != nil {
		return err
	}

	if fieldType.OmitNorms() {
		err := p.fieldInfo.SetOmitsNorms()
		if err != nil {
			return err
		}
	}

	analyzed := fieldType.Tokenized() && p.analyzer != nil

	// To assist people in tracking down problems in analysis components, we wish to write the field name to the infostream
	// when we fail. We expect some caller to eventually deal with the real exception, so we don't want any 'catch' clauses,
	// but rather a finally that takes note of the problem.
	//succeededInProcessingField := false

	stream, err := field.TokenStream(p.analyzer, p.tokenStream)
	if err != nil {
		return err
	}
	// reset the TokenStream to the first token
	if err := stream.Reset(); err != nil {
		return err
	}

	p.invertState.SetAttributeSource(stream.AttributeSource())
	p.termsHashPerField.Start(field, first)

	for {
		ok, err := stream.IncrementToken()
		if err != nil {
			return err
		}

		if !ok {
			break
		}

		// If we hit an exception in stream.next below
		// (which is fairly common, e.g. if analyzer
		// chokes on a given document), then it's
		// non-aborting and (above) this one document
		// will be marked as deleted, but still
		// consume a docID

		posIncr := p.invertState.posIncrAttribute.GetPositionIncrement()
		p.invertState.position += posIncr
		if p.invertState.position < p.invertState.lastPosition {
			if posIncr == 0 {
				return fmt.Errorf("first position increment must be > 0 (got 0) for field '%s'", field.Name())
			} else if posIncr < 0 {
				return fmt.Errorf("position increment must be >= 0 (got %d) for field '%s'", posIncr, field.Name())
			} else {
				return fmt.Errorf("position overflowed Integer.MAX_VALUE")
			}
		} else if p.invertState.position > MAX_POSITION {
			return fmt.Errorf("position %s is too large for field '': max allowed position is %d", field.Name(), MAX_POSITION)
		}
		p.invertState.lastPosition = p.invertState.position
		if posIncr == 0 {
			p.invertState.numOverlap++
		}

		startOffset := p.invertState.offset + p.invertState.offsetAttribute.StartOffset()
		endOffset := p.invertState.offset + p.invertState.offsetAttribute.EndOffset()
		if startOffset < p.invertState.lastStartOffset || endOffset < startOffset {
			return fmt.Errorf("startOffset must be non-negative, and endOffset must" +
				" be >= startOffset, and offsets must not go backwards ")
		}
		p.invertState.lastStartOffset = startOffset
		// TODO: fix overlap
		p.invertState.length += p.invertState.termFreqAttribute.GetTermFrequency()

		// If we hit an exception in here, we abort
		// all buffered documents since the last
		// Flush, on the likelihood that the
		// internal state of the terms hash is now
		// corrupt and should not be flushed to a
		// new segment:
		if err := p.termsHashPerField.Add(p.invertState.termAttribute.GetBytes(), docID); err != nil {
			return err
		}
	}

	// trigger streams to perform end-of-stream operations
	if err := stream.End(); err != nil {
		return err
	}

	// TODO: maybe add some safety? then again, it's already checked
	// when we come back around to the field...
	p.invertState.position += p.invertState.posIncrAttribute.GetPositionIncrement()
	p.invertState.offset += p.invertState.offsetAttribute.EndOffset()

	/* if there is an exception coming through, we won't set this to true here:*/
	//succeededInProcessingField = true

	if analyzed {
		p.invertState.position += p.analyzer.GetPositionIncrementGap(p.fieldInfo.Name())
		p.invertState.offset += p.analyzer.GetOffsetGap(p.fieldInfo.Name())
	}
	return nil
}

func (p *PerField) Finish(docID int) error {
	if p.fieldInfo.OmitsNorms() == false {

		// the field exists in this document, but it did not have
		// any indexed tokens, so we assign a default item of zero
		// to the norm
		normValue := int64(0)
		if p.invertState.length != 0 {
			normValue = p.similarity.ComputeNorm(p.invertState)
			if normValue == 0 {
				return errors.New("return 0 for no-empty field")
				//throw new IllegalStateException("Similarity " + similarity + " return 0 for non-empty field");
			}
		}
		p.norms.AddValue(docID, normValue)
	}

	return p.termsHashPerField.Finish()
}

func newIntBlockAllocator() ints.IntsAllocator {
	return &ints.IntsAllocatorDefault{
		BlockSize: ints.INT_BLOCK_SIZE,
		FnRecycleIntBlocks: func(blocks [][]int, start, end int) {
			return
		},
	}
}

func newByteBlockAllocator() bytesref.Allocator {
	fn := func(blocks [][]byte, start, end int) {
		for i := start; i < end; i++ {
			blocks[i] = nil
		}
	}
	return bytesref.GetAllocatorBuilder().NewBytes(bytesref.BYTE_BLOCK_SIZE, fn)
}
