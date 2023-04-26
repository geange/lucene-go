package index

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/util"
)

var _ LeafReader = &DocValuesLeafReader{}

type DocValuesLeafReader struct {
	*LeafReaderDefault
}

func NewDocValuesLeafReader() *DocValuesLeafReader {
	reader := &DocValuesLeafReader{}

	reader.LeafReaderDefault = NewLeafReaderDefault(reader)
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

func (d *DocValuesLeafReader) GetLiveDocs() util.Bits {
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
