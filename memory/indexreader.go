package memory

import (
	"context"
	"math"

	"github.com/geange/lucene-go/core/interface/index"

	"github.com/geange/gods-generic/maps/treemap"
	"github.com/geange/lucene-go/core/document"
	cindex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/bytesref"
	"github.com/geange/lucene-go/core/util/version"
)

var _ index.LeafReader = &IndexReader{}

// IndexReader
// Search support for Lucene framework integration; implements all methods required by the Lucene
// Reader contracts.
type IndexReader struct {
	*cindex.BaseLeafReader

	memoryFields *Fields
	fieldInfos   index.FieldInfos
	fields       *treemap.Map[string, *info]
}

func newIndexReader(memoryFields *Fields, fieldInfos index.FieldInfos,
	fields *treemap.Map[string, *info]) *IndexReader {

	newReader := &IndexReader{
		memoryFields: memoryFields,
		fieldInfos:   fieldInfos,
		fields:       fields,
	}

	leafReader := cindex.NewBaseLeafReader(newReader)
	newReader.BaseLeafReader = leafReader
	return newReader
}

func (m *IndexReader) GetTermVectors(docID int) (index.Fields, error) {
	if docID == 0 {
		return m.memoryFields, nil
	} else {
		return nil, nil
	}
}

func (m *IndexReader) NumDocs() int {
	return 1
}

func (m *IndexReader) MaxDoc() int {
	return 1
}

func (m *IndexReader) DocumentWithVisitor(ctx context.Context, docID int, visitor document.StoredFieldVisitor) error {
	return nil
}

func (m *IndexReader) DoClose() error {
	return nil
}

func (m *IndexReader) GetReaderCacheHelper() index.CacheHelper {
	return nil
}

func (m *IndexReader) Terms(field string) (index.Terms, error) {
	return m.memoryFields.Terms(field)
}

func (m *IndexReader) GetNumericDocValues(field string) (index.NumericDocValues, error) {
	info := m.getInfoForExpectedDocValuesType(field, document.DOC_VALUES_TYPE_NUMERIC)
	if info != nil {
		return newNumericDocValues(int64(info.numericProducer.dvLongValues[0])), nil
	}
	return nil, nil
}

func (m *IndexReader) GetBinaryDocValues(field string) (index.BinaryDocValues, error) {
	return m.getSortedDocValues(field, document.DOC_VALUES_TYPE_SORTED), nil
}

func (m *IndexReader) GetSortedDocValues(field string) (index.SortedDocValues, error) {
	return m.getSortedDocValues(field, document.DOC_VALUES_TYPE_SORTED), nil
}

func (m *IndexReader) getSortedDocValues(field string, docValuesType document.DocValuesType) index.SortedDocValues {
	info := m.getInfoForExpectedDocValuesType(field, document.DOC_VALUES_TYPE_SORTED_SET)
	if info != nil {
		value := info.binaryProducer.dvBytesValuesSet.Get(0)
		return newSortedDocValues(value)
	}
	return nil
}

func (m *IndexReader) GetSortedNumericDocValues(field string) (index.SortedNumericDocValues, error) {
	info := m.getInfoForExpectedDocValuesType(field, document.DOC_VALUES_TYPE_SORTED_NUMERIC)
	if info != nil {
		values := info.numericProducer.dvLongValues
		count := info.numericProducer.count
		return newSortedNumericDocValues(values, count), nil
	}
	return nil, nil
}

func (m *IndexReader) GetNormValues(field string) (index.NumericDocValues, error) {
	info, ok := m.fields.Get(field)
	if !ok {
		return nil, nil
	}

	if info.fieldInfo.OmitsNorms() {
		return nil, nil
	}

	return info.getNormDocValues(), nil
}

func (m *IndexReader) GetSortedSetDocValues(field string) (index.SortedSetDocValues, error) {
	info := m.getInfoForExpectedDocValuesType(field, document.DOC_VALUES_TYPE_SORTED_SET)
	if info != nil {
		return genSortedSetDocValues(info.terms, info.binaryProducer.bytesIds), nil
	}
	return nil, nil
}

func (m *IndexReader) getInfoForExpectedDocValuesType(fieldName string, expectedType document.DocValuesType) *info {
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

func (m *IndexReader) GetFieldInfos() index.FieldInfos {
	return m.fieldInfos
}

func (m *IndexReader) GetLiveDocs() util.Bits {
	return nil
}

func (m *IndexReader) GetPointValues(field string) (types.PointValues, bool) {
	info, ok := m.fields.Get(field)
	if ok {
		return newMemoryIndexPointValues(info), true
	}
	return nil, false
}

func (m *IndexReader) CheckIntegrity() error {
	return nil
}

func (m *IndexReader) GetMetaData() index.LeafMetaData {
	return cindex.NewLeafMetaData(int(version.Last.Major()), version.Last, nil)
}

func genSortedSetDocValues(values *bytesref.BytesHash, bytesIds []int) index.SortedSetDocValues {
	it := newDocValuesIterator()
	return newSortedSetDocValues(values, bytesIds, it)
}

type docValuesIterator struct {
	doc int
}

func newDocValuesIterator() *docValuesIterator {
	return &docValuesIterator{doc: -1}
}

func (r *docValuesIterator) advance(doc int) (int, error) {
	r.doc = doc
	return r.docId()
}

func (r *docValuesIterator) nextDoc() (int, error) {
	r.doc++
	return r.docId()
}

func (r *docValuesIterator) docId() (int, error) {
	if r.doc > 0 {
		return r.doc, nil
	}
	return math.MaxInt32, nil
}
