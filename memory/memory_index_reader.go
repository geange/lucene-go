package memory

import (
	"github.com/geange/lucene-go/core/types"
	"math"

	"github.com/geange/gods-generic/maps/treemap"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util"
)

var _ index.LeafReader = &MemoryIndexReader{}

// MemoryIndexReader Search support for Lucene framework integration; implements all methods required by the Lucene
// Reader contracts.
type MemoryIndexReader struct {
	*index.LeafReaderBase

	memoryFields *MemoryFields
	fieldInfos   *index.FieldInfos

	fields *treemap.Map[string, *Info]

	*MemoryIndex
}

func (m *MemoryIndex) NewMemoryIndexReader(fields *treemap.Map[string, *Info]) *MemoryIndexReader {
	fieldInfosArr := make([]*document.FieldInfo, fields.Size())
	i := 0
	it := fields.Iterator()

	for it.Next() {
		info := it.Value()
		info.prepareDocValuesAndPointValues()
		fieldInfosArr[i] = info.fieldInfo
		i++
	}

	reader := &MemoryIndexReader{
		memoryFields: m.NewMemoryFields(fields),
		fieldInfos:   index.NewFieldInfos(fieldInfosArr),
		fields:       fields,
	}

	leafReader := index.NewLeafReaderDefault(reader)
	reader.LeafReaderBase = leafReader

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

func (m *MemoryIndexReader) DocumentV1(docID int, visitor document.StoredFieldVisitor) error {
	return nil
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
	info := m.getInfoForExpectedDocValuesType(field, document.DOC_VALUES_TYPE_NUMERIC)
	if info != nil {
		return newInnerNumericDocValues(int64(info.numericProducer.dvLongValues[0])), nil
	}
	return nil, nil
}

func (m *MemoryIndexReader) GetBinaryDocValues(field string) (index.BinaryDocValues, error) {
	return m.getSortedDocValues(field, document.DOC_VALUES_TYPE_SORTED), nil
}

func (m *MemoryIndexReader) GetSortedDocValues(field string) (index.SortedDocValues, error) {
	return m.getSortedDocValues(field, document.DOC_VALUES_TYPE_SORTED), nil
}

func (m *MemoryIndexReader) getSortedDocValues(field string, docValuesType document.DocValuesType) index.SortedDocValues {
	info := m.getInfoForExpectedDocValuesType(field, document.DOC_VALUES_TYPE_SORTED_SET)
	if info != nil {
		value := info.binaryProducer.dvBytesValuesSet.Get(0)
		return newInnerSortedDocValues(value)
	}
	return nil
}

func (m *MemoryIndexReader) GetSortedNumericDocValues(field string) (index.SortedNumericDocValues, error) {
	info := m.getInfoForExpectedDocValuesType(field, document.DOC_VALUES_TYPE_SORTED_NUMERIC)
	if info != nil {
		return newInnerSortedNumericDocValues(info.numericProducer.dvLongValues, info.numericProducer.count), nil
	}
	return nil, nil
}

func (m *MemoryIndexReader) GetNormValues(field string) (index.NumericDocValues, error) {
	info, ok := m.fields.Get(field)
	if !ok {
		return nil, nil
	}

	if info.fieldInfo.OmitsNorms() {
		return nil, nil
	}

	return info.getNormDocValues(), nil
}

func (m *MemoryIndexReader) GetSortedSetDocValues(field string) (index.SortedSetDocValues, error) {
	info := m.getInfoForExpectedDocValuesType(field, document.DOC_VALUES_TYPE_SORTED_SET)
	if info != nil {
		return sortedSetDocValues(info.terms, info.binaryProducer.bytesIds), nil
	}
	return nil, nil
}

func (m *MemoryIndexReader) getInfoForExpectedDocValuesType(fieldName string, expectedType document.DocValuesType) *Info {
	if expectedType == document.DOC_VALUES_TYPE_NONE {
		return nil
	}

	info, found := m.fields.Get(fieldName)
	if !found {
		return nil
	}

	if info.fieldInfo.GetDocValuesType() != expectedType {
		return nil
	}
	return info
}

func (m *MemoryIndexReader) GetFieldInfos() *index.FieldInfos {
	return m.fieldInfos
}

func (m *MemoryIndexReader) GetLiveDocs() util.Bits {
	return nil
}

func (m *MemoryIndexReader) GetPointValues(field string) (types.PointValues, bool) {
	info, ok := m.fields.Get(field)
	if ok {
		return newMemoryIndexPointValues(info), true
	}
	return nil, false
}

func (m *MemoryIndexReader) CheckIntegrity() error {
	return nil
}

func (m *MemoryIndexReader) GetMetaData() *index.LeafMetaData {
	return index.NewLeafMetaData(util.VersionLast.Major, util.VersionLast, nil)
}

func sortedSetDocValues(values *util.BytesHash, bytesIds []int) index.SortedSetDocValues {
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
