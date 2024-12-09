package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/geange/lucene-go/core/interface/index"

	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/bytesref"
	"github.com/geange/lucene-go/core/util/ints"
)

var _ index.DocConsumer = &DefaultIndexingChain{}

// DefaultIndexingChain
// Default general purpose indexing chain, which handles indexing all types of fields.
// 默认的通用索引链，用于处理所有类型的字段的索引。
type DefaultIndexingChain struct {
	fieldInfos               *FieldInfosBuilder
	termsHash                TermsHash
	docValuesBytePool        *bytesref.BlockPool
	storedFieldsConsumer     *StoredFieldsConsumer
	termVectorsWriter        *TermVectorsConsumer
	fieldHash                map[string]*PerField // 使用map简化理解
	totalFieldCount          int
	nextFieldGen             int64
	fields                   []*PerField
	byteBlockAllocator       bytesref.Allocator
	indexWriterConfig        *liveIndexWriterConfig
	indexCreatedVersionMajor int
	hasHitAbortingException  bool
}

func NewDefaultIndexingChain(indexCreatedVersionMajor int, segmentInfo *SegmentInfo, dir store.Directory,
	fieldInfos *FieldInfosBuilder, indexWriterConfig *liveIndexWriterConfig) *DefaultIndexingChain {

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

func (d *DefaultIndexingChain) getDocValuesLeafReader() index.LeafReader {
	reader := &defaultIndexingChainLeafReader{
		DocValuesLeafReader: NewDocValuesLeafReader(),
		chain:               d,
	}
	return reader
}

var _ index.LeafReader = &defaultIndexingChainLeafReader{}

type defaultIndexingChainLeafReader struct {
	*DocValuesLeafReader

	chain *DefaultIndexingChain
}

func (r *defaultIndexingChainLeafReader) GetNumericDocValues(field string) (index.NumericDocValues, error) {
	pf := r.chain.getPerField(field)
	if pf == nil {
		return nil, nil
	}

	if pf.fieldInfo.GetDocValuesType() == document.DOC_VALUES_TYPE_NUMERIC {
		docValues := pf.docValuesWriter.GetDocValues()
		numericDocValues, ok := docValues.(index.NumericDocValues)
		if !ok {
			return nil, errors.New("expected numeric doc values")
		}
		return numericDocValues, nil
	}
	return nil, nil
}

func (r *defaultIndexingChainLeafReader) GetBinaryDocValues(field string) (index.BinaryDocValues, error) {
	pf := r.chain.getPerField(field)
	if pf == nil {
		return nil, nil
	}
	if pf.fieldInfo.GetDocValuesType() == document.DOC_VALUES_TYPE_BINARY {
		return pf.docValuesWriter.GetDocValues().(index.BinaryDocValues), nil
	}
	return nil, nil
}

func (r *defaultIndexingChainLeafReader) GetSortedDocValues(field string) (index.SortedDocValues, error) {
	pf := r.chain.getPerField(field)
	if pf == nil {
		return nil, nil
	}
	if pf.fieldInfo.GetDocValuesType() == document.DOC_VALUES_TYPE_SORTED {
		return pf.docValuesWriter.GetDocValues().(index.SortedDocValues), nil
	}
	return nil, nil
}

func (r *defaultIndexingChainLeafReader) GetSortedNumericDocValues(field string) (index.SortedNumericDocValues, error) {
	pf := r.chain.getPerField(field)
	if pf == nil {
		return nil, nil
	}
	if pf.fieldInfo.GetDocValuesType() == document.DOC_VALUES_TYPE_SORTED_NUMERIC {
		docValues := pf.docValuesWriter.GetDocValues()
		sortedNumericDocValues, ok := docValues.(index.SortedNumericDocValues)
		if !ok {
			return nil, errors.New("docValues is not SortedNumericDocValues")
		}
		return sortedNumericDocValues, nil
	}
	return nil, nil
}

func (r *defaultIndexingChainLeafReader) GetSortedSetDocValues(field string) (index.SortedSetDocValues, error) {
	pf := r.chain.getPerField(field)
	if pf == nil {
		return nil, nil
	}
	if pf.fieldInfo.GetDocValuesType() == document.DOC_VALUES_TYPE_SORTED_SET {
		return pf.docValuesWriter.GetDocValues().(index.SortedSetDocValues), nil
	}
	return nil, nil
}

func (r *defaultIndexingChainLeafReader) GetFieldInfos() index.FieldInfos {
	return r.chain.fieldInfos.Finish()
}

func (d *DefaultIndexingChain) maybeSortSegment(state *index.SegmentWriteState) (index.DocMap, error) {
	indexSort := state.SegmentInfo.GetIndexSort()
	if indexSort == nil {
		return nil, nil
	}
	docValuesReader := d.getDocValuesLeafReader()
	comparators := make([]index.DocComparator, 0)

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

func (d *DefaultIndexingChain) Flush(ctx context.Context, state *index.SegmentWriteState) (index.DocMap, error) {
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
	if err := d.writePoints(ctx, state, sortMap); err != nil {
		return nil, err
	}

	if err := d.storedFieldsConsumer.Finish(ctx, maxDoc); err != nil {
		return nil, err
	}
	if err := d.storedFieldsConsumer.Flush(ctx, state, sortMap); err != nil {
		return nil, err
	}

	fieldsToFlush := make(map[string]TermsHashPerField)

	for _, perField := range d.fieldHash {
		fieldsToFlush[perField.fieldInfo.Name()] = perField.termsHashPerField
	}

	readState := index.NewSegmentReadState(state.Directory, state.SegmentInfo, state.FieldInfos, state.Context, state.SegmentSuffix)

	//var norms NormsProducer
	if readState.FieldInfos.HasNorms() {
		norms, err := state.SegmentInfo.GetCodec().NormsFormat().NormsProducer(ctx, readState)
		if err != nil {
			return nil, err
		}

		normsMergeInstance := norms.GetMergeInstance()
		if err := d.termsHash.Flush(nil, fieldsToFlush, state, sortMap, normsMergeInstance); err != nil {
			return nil, err
		}
	}

	if err := d.indexWriterConfig.GetCodec().FieldInfosFormat().
		Write(ctx, state.Directory, state.SegmentInfo, "", state.FieldInfos, state.Context); err != nil {
		return nil, err
	}
	return sortMap, nil
}

// Writes all buffered points.
func (d *DefaultIndexingChain) writePoints(ctx context.Context, state *index.SegmentWriteState, sortMap index.DocMap) error {
	var pointsWriter index.PointsWriter
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
				pointsWriter, err = format.FieldsWriter(nil, state)
				if err != nil {
					return err
				}
			}

			if err := perField.pointValuesWriter.Flush(ctx, state, sortMap, pointsWriter); err != nil {
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
func (d *DefaultIndexingChain) writeDocValues(state *index.SegmentWriteState, sortMap index.DocMap) error {
	var dvConsumer index.DocValuesConsumer
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
				dvConsumer, err = format.FieldsConsumer(nil, state)
				if err != nil {
					return err
				}
			}
			if err := perField.docValuesWriter.Flush(state, sortMap, dvConsumer); err != nil {
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

func (d *DefaultIndexingChain) writeNorms(state *index.SegmentWriteState, sortMap index.DocMap) error {
	var normsConsumer index.NormsConsumer
	var err error
	if state.FieldInfos.HasNorms() {
		normsFormat := state.SegmentInfo.GetCodec().NormsFormat()
		normsConsumer, err = normsFormat.NormsConsumer(nil, state)
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
	if err := d.startStoredFields(ctx, docId); err != nil {
		return err
	}

	for _, field := range doc.Fields() {
		count, err := d.processField(ctx, docId, field, fieldGen, fieldCount)
		if err != nil {
			return err
		}
		fieldCount = count

		for i := 0; i < fieldCount; i++ {
			if err := d.fields[i].Finish(docId); err != nil {
				return err
			}
		}
		if err := d.finishStoredFields(); err != nil {
			return err
		}
	}

	return d.termsHash.FinishDocument(nil, docId)
}

func (d *DefaultIndexingChain) processField(ctx context.Context, docId int, field document.IndexableField, fieldGen int64, fieldCount int) (int, error) {

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
		isFirst := fp.fieldGen != fieldGen
		if err := fp.invert(docId, field, isFirst); err != nil {
			return 0, err
		}

		if isFirst {
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
			if err := d.storedFieldsConsumer.writeField(ctx, fp.fieldInfo, field); err != nil {
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

func (d *DefaultIndexingChain) validateIndexSortDVType(indexSort index.Sort, fieldToValidate string, dvType document.DocValuesType) error {
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

func (d *DefaultIndexingChain) startStoredFields(ctx context.Context, docID int) error {
	return d.storedFieldsConsumer.StartDocument(ctx, docID)
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
	similarity               index.Similarity
	invertState              *index.FieldInvertState
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
	invert bool, similarity index.Similarity, analyzer analysis.Analyzer) (*PerField, error) {

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
	p.invertState = index.NewFieldInvertStateV1(
		p.indexCreatedVersionMajor, p.fieldInfo.Name(), p.fieldInfo.GetIndexOptions())

	termsHashPerField, err := p.chain.termsHash.AddField(p.invertState, p.fieldInfo)
	if err != nil {
		return err
	}
	p.termsHashPerField = termsHashPerField

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
		if err := p.fieldInfo.SetOmitsNorms(); err != nil {
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

		posIncr := p.invertState.PosIncrAttribute.GetPositionIncrement()
		p.invertState.Position += posIncr
		if p.invertState.Position < p.invertState.LastPosition {
			if posIncr == 0 {
				return fmt.Errorf("first position increment must be > 0 (got 0) for field '%s'", field.Name())
			} else if posIncr < 0 {
				return fmt.Errorf("position increment must be >= 0 (got %d) for field '%s'", posIncr, field.Name())
			} else {
				return fmt.Errorf("position overflowed Integer.MAX_VALUE")
			}
		} else if p.invertState.Position > MAX_POSITION {
			return fmt.Errorf("position %s is too large for field '': max allowed position is %d", field.Name(), MAX_POSITION)
		}
		p.invertState.LastPosition = p.invertState.Position
		if posIncr == 0 {
			p.invertState.NumOverlap++
		}

		startOffset := p.invertState.Offset + p.invertState.OffsetAttribute.StartOffset()
		endOffset := p.invertState.Offset + p.invertState.OffsetAttribute.EndOffset()
		if startOffset < p.invertState.LastStartOffset || endOffset < startOffset {
			return fmt.Errorf("startOffset must be non-negative, and endOffset must" +
				" be >= startOffset, and offsets must not go backwards ")
		}
		p.invertState.LastStartOffset = startOffset
		// TODO: fix overlap
		p.invertState.Length += p.invertState.TermFreqAttribute.GetTermFrequency()

		// If we hit an exception in here, we abort
		// all buffered documents since the last
		// Flush, on the likelihood that the
		// internal state of the terms hash is now
		// corrupt and should not be flushed to a
		// new segment:
		if err := p.termsHashPerField.Add(p.invertState.TermAttribute.GetBytes(), docID); err != nil {
			return err
		}
	}

	// trigger streams to perform end-of-stream operations
	if err := stream.End(); err != nil {
		return err
	}

	// TODO: maybe add some safety? then again, it's already checked
	// when we come back around to the field...
	p.invertState.Position += p.invertState.PosIncrAttribute.GetPositionIncrement()
	p.invertState.Offset += p.invertState.OffsetAttribute.EndOffset()

	/* if there is an exception coming through, we won't set this to true here:*/
	//succeededInProcessingField = true

	if analyzed {
		p.invertState.Position += p.analyzer.GetPositionIncrementGap(p.fieldInfo.Name())
		p.invertState.Offset += p.analyzer.GetOffsetGap(p.fieldInfo.Name())
	}
	return nil
}

func (p *PerField) Finish(docID int) error {
	if p.fieldInfo.OmitsNorms() == false {

		// the field exists in this document, but it did not have
		// any indexed tokens, so we assign a default item of zero
		// to the norm
		normValue := int64(0)
		if p.invertState.Length != 0 {
			normValue = p.similarity.ComputeNorm(p.invertState)
			if normValue == 0 {
				return errors.New("return 0 for no-empty field")
				//throw new IllegalStateException("Similarity " + similarity + " return 0 for non-empty field");
			}
		}
		if err := p.norms.AddValue(docID, normValue); err != nil {
			return err
		}
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
