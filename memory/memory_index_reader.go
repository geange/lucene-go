package memory

import (
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
	memoryFields *MemoryFields
	fieldInfos   *index.FieldInfos

	fields *treemap.Map
}

func (m *MemoryIndexReader) Close() error {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) GetTermVectors(docID int) (index.Fields, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) GetTermVector(docID int, field string) (index.Terms, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) NumDocs() int {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) MaxDoc() int {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) NumDeletedDocs() int {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) Document(docID int) (*document.Document, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) DocumentV1(docID int, visitor document.StoredFieldVisitor) (*document.Document, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) DocumentV2(docID int, fieldsToLoad map[string]struct{}) (*document.Document, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) HasDeletions() bool {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) DoClose() error {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) GetContext() index.IndexReaderContext {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) Leaves() ([]index.LeafReaderContext, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) GetReaderCacheHelper() index.CacheHelper {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) DocFreq(term index.Term) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) TotalTermFreq(term *index.Term) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) GetSumDocFreq(field string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) GetDocCount(field string) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) GetSumTotalTermFreq(field string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) Terms(field string) (index.Terms, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) Postings(term *index.Term, flags int) (index.PostingsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) GetSortedSetDocValues(field string) (index.SortedSetDocValues, error) {
	info := m.getInfoForExpectedDocValuesType(field, types.DOC_VALUES_TYPE_SORTED_SET)
	if info != nil {
		return sortedSetDocValues(info.terms, info.binaryProducer.bytesIds), nil
	}
	return nil, nil
}

func (m *MemoryIndexReader) getInfoForExpectedDocValuesType(fieldName string, expectedType types.DocValuesType) *Info {
	panic("")
}

func (m *MemoryIndexReader) GetFieldInfos() *index.FieldInfos {
	return m.fieldInfos
}

func (m *MemoryIndexReader) GetLiveDocs() util.Bits {
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
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexReader) GetMetaData() *index.LeafMetaData {
	//TODO implement me
	panic("implement me")
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
