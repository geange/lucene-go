package memory

import (
	"github.com/bits-and-blooms/bitset"
	"github.com/emirpasic/gods/maps/treemap"
	_ "github.com/emirpasic/gods/maps/treemap"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"math"
)

var _ index.LeafReader = &MemoryIndexReader{}

// MemoryIndexReader Search support for Lucene framework integration; implements all methods required by the Lucene
// IndexReader contracts.
type MemoryIndexReader struct {
	*index.LeafReaderImp

	memoryFields *MemoryFields
	fieldInfos   *index.FieldInfos

	fields *treemap.Map
}

func NewMemoryIndexReader(fields *treemap.Map) *MemoryIndexReader {
	fieldInfosArr := make([]types.FieldInfo, fields.Size())
	i := 0
	it := fields.Iterator()

	for it.Next() {
		info := it.Value().(*Info)
		info.prepareDocValuesAndPointValues()
		fieldInfosArr[i] = *info.fieldInfo
		i++
	}

	reader := &MemoryIndexReader{
		LeafReaderImp: nil,
		memoryFields:  NewMemoryFields(fields),
		fieldInfos:    index.NewFieldInfos(fieldInfosArr),
		fields:        fields,
	}

	reader.LeafReaderImp = index.NewLeafReaderImp(reader, reader, index.NewLeafReaderContext(reader))
	return reader
}

func (m *MemoryIndexReader) GetTermVectors(docID int) (index.Fields, error) {
	if docID == 0 {
		return m.memoryFields, nil
	} else {
		return nil, nil
	}
}

func (m *MemoryIndexReader) NumDocs() int {
	return 1
}

func (m *MemoryIndexReader) MaxDoc() int {
	return 1
}

func (m *MemoryIndexReader) DocumentV1(docID int, visitor document.StoredFieldVisitor) (*document.Document, error) {
	return nil, nil
}

func (m *MemoryIndexReader) DoClose() error {
	return nil
}

func (m *MemoryIndexReader) GetReaderCacheHelper() index.CacheHelper {
	return nil
}

func (m *MemoryIndexReader) Terms(field string) (index.Terms, error) {
	return m.memoryFields.Terms(field)
}

func (m *MemoryIndexReader) GetNumericDocValues(field string) (index.NumericDocValues, error) {
	info := m.getInfoForExpectedDocValuesType(field, types.DOC_VALUES_TYPE_NUMERIC)
	if info != nil {
		return newInnerNumericDocValues(int64(info.numericProducer.dvLongValues[0])), nil
	}
	return nil, nil
}

func (m *MemoryIndexReader) GetBinaryDocValues(field string) (index.BinaryDocValues, error) {
	return m.getSortedDocValues(field, types.DOC_VALUES_TYPE_SORTED), nil
}

func (m *MemoryIndexReader) GetSortedDocValues(field string) (index.SortedDocValues, error) {
	return m.getSortedDocValues(field, types.DOC_VALUES_TYPE_SORTED), nil
}

func (m *MemoryIndexReader) getSortedDocValues(field string, docValuesType types.DocValuesType) index.SortedDocValues {
	info := m.getInfoForExpectedDocValuesType(field, types.DOC_VALUES_TYPE_SORTED_SET)
	if info != nil {
		value := info.binaryProducer.dvBytesValuesSet.Get(0)
		return newInnerSortedDocValues(value)
	}
	return nil
}

func (m *MemoryIndexReader) GetSortedNumericDocValues(field string) (index.SortedNumericDocValues, error) {
	info := m.getInfoForExpectedDocValuesType(field, types.DOC_VALUES_TYPE_SORTED_NUMERIC)
	if info != nil {
		return newInnerSortedNumericDocValues(info.numericProducer.dvLongValues, info.numericProducer.count), nil
	}
	return nil, nil
}

func (m *MemoryIndexReader) GetNormValues(field string) (index.NumericDocValues, error) {
	v, _ := m.fields.Get(field)
	info := v.(*Info)
	if info == nil {
		return nil, nil
	}

	info, ok := v.(*Info)
	if !ok {
		return nil, nil
	}

	if info.fieldInfo.OmitsNorms() {
		return nil, nil
	}

	return info.getNormDocValues(), nil
}

func (m *MemoryIndexReader) GetSortedSetDocValues(field string) (index.SortedSetDocValues, error) {
	info := m.getInfoForExpectedDocValuesType(field, types.DOC_VALUES_TYPE_SORTED_SET)
	if info != nil {
		return sortedSetDocValues(info.terms, info.binaryProducer.bytesIds), nil
	}
	return nil, nil
}

func (m *MemoryIndexReader) getInfoForExpectedDocValuesType(fieldName string, expectedType types.DocValuesType) *Info {
	if expectedType == types.DOC_VALUES_TYPE_NONE {
		return nil
	}

	v, found := m.fields.Get(fieldName)
	if !found {
		return nil
	}

	info := v.(*Info)
	if info.fieldInfo.GetDocValuesType() != expectedType {
		return nil
	}
	return info
}

func (m *MemoryIndexReader) GetFieldInfos() *index.FieldInfos {
	return m.fieldInfos
}

func (m *MemoryIndexReader) GetLiveDocs() *bitset.BitSet {
	return nil
}

func (m *MemoryIndexReader) GetPointValues(field string) (index.PointValues, error) {
	v, ok := m.fields.Get(field)
	if ok {
		info := v.(*Info)
		return newMemoryIndexPointValues(info), nil
	}
	return nil, nil
}

func (m *MemoryIndexReader) CheckIntegrity() error {
	return nil
}

func (m *MemoryIndexReader) GetMetaData() *index.LeafMetaData {
	return index.NewLeafMetaData(util.VersionLast.Major, util.VersionLast)
}

func sortedSetDocValues(values *util.BytesRefHash, bytesIds []int) index.SortedSetDocValues {
	it := newMemoryDocValuesIterator()
	return newInnerSortedSetDocValues(values, bytesIds, it)
}

type memoryDocValuesIterator struct {
	doc int
}

func newMemoryDocValuesIterator() *memoryDocValuesIterator {
	return &memoryDocValuesIterator{doc: -1}
}

func (r *memoryDocValuesIterator) advance(doc int) int {
	r.doc = doc
	return r.docId()
}

func (r *memoryDocValuesIterator) nextDoc() int {
	r.doc++
	return r.docId()
}

func (r *memoryDocValuesIterator) docId() int {
	if r.doc > 0 {
		return r.doc
	}
	return math.MaxInt32
}
