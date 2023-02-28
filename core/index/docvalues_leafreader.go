package index

import (
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/document"
)

var _ LeafReader = &DocValuesLeafReader{}

type DocValuesLeafReader struct {
	*LeafReaderDefault
}

func NewDocValuesLeafReader() *DocValuesLeafReader {
	reader := &DocValuesLeafReader{}

	reader.LeafReaderDefault = NewLeafReaderDefault(&LeafReaderDefaultConfig{
		Terms:         reader.Terms,
		ReaderContext: nil,
		IndexReaderDefaultConfig: IndexReaderDefaultConfig{
			GetTermVectors:       reader.GetTermVectors,
			NumDocs:              reader.NumDocs,
			MaxDoc:               reader.MaxDoc,
			DocumentV1:           reader.DocumentV1,
			GetContext:           reader.GetContext,
			DoClose:              reader.DoClose,
			GetReaderCacheHelper: reader.GetReaderCacheHelper,
			DocFreq:              reader.DocFreq,
			TotalTermFreq:        reader.TotalTermFreq,
			GetSumDocFreq:        reader.GetSumDocFreq,
			GetDocCount:          reader.GetDocCount,
			GetSumTotalTermFreq:  reader.GetSumTotalTermFreq,
		},
	})
	return reader
}

func (d *DocValuesLeafReader) GetTermVectors(docID int) (Fields, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) NumDocs() int {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) MaxDoc() int {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) DocumentV1(docID int, visitor document.StoredFieldVisitor) error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) DoClose() error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) GetReaderCacheHelper() CacheHelper {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) Terms(field string) (Terms, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) GetNumericDocValues(field string) (NumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) GetBinaryDocValues(field string) (BinaryDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) GetSortedDocValues(field string) (SortedDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) GetSortedNumericDocValues(field string) (SortedNumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) GetSortedSetDocValues(field string) (SortedSetDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) GetNormValues(field string) (NumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) GetFieldInfos() *FieldInfos {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) GetLiveDocs() *bitset.BitSet {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) GetPointValues(field string) (PointValues, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (d *DocValuesLeafReader) GetMetaData() *LeafMetaData {
	//TODO implement me
	panic("implement me")
}
