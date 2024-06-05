package index

import (
	"errors"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

var _ LeafReader = &DocValuesLeafReader{}

type DocValuesLeafReader struct {
	*BaseLeafReader
}

func NewDocValuesLeafReader() *DocValuesLeafReader {
	reader := &DocValuesLeafReader{}

	reader.BaseLeafReader = NewBaseLeafReader(reader)
	return reader
}

func (d *DocValuesLeafReader) GetTermVectors(docID int) (Fields, error) {
	return nil, errors.New("GetTermVectors is not yet implemented")
}

func (d *DocValuesLeafReader) NumDocs() int {
	return 0
}

func (d *DocValuesLeafReader) MaxDoc() int {
	return 0
}

func (d *DocValuesLeafReader) DocumentWithVisitor(docID int, visitor document.StoredFieldVisitor) error {
	return errors.New("func DocumentWithVisitor is not yet implemented")
}

func (d *DocValuesLeafReader) DoClose() error {
	return errors.New("funcDoClose is not yet implemented")
}

func (d *DocValuesLeafReader) GetReaderCacheHelper() CacheHelper {
	return nil
}

func (d *DocValuesLeafReader) Terms(field string) (Terms, error) {
	return nil, errors.New("func Terms is not yet implemented")
}

func (d *DocValuesLeafReader) GetNumericDocValues(field string) (NumericDocValues, error) {
	return nil, errors.New("func NumericDocValues is not yet implemented")
}

func (d *DocValuesLeafReader) GetBinaryDocValues(field string) (BinaryDocValues, error) {
	return nil, errors.New("func BinaryDocValues is not yet implemented")
}

func (d *DocValuesLeafReader) GetSortedDocValues(field string) (SortedDocValues, error) {
	return nil, errors.New("func SortedDocValues is not yet implemented")
}

func (d *DocValuesLeafReader) GetSortedNumericDocValues(field string) (SortedNumericDocValues, error) {
	return nil, errors.New("func SortedNumericDocValues is not yet implemented")
}

func (d *DocValuesLeafReader) GetSortedSetDocValues(field string) (SortedSetDocValues, error) {
	return nil, errors.New("func SortedSetDocValues is not yet implemented")
}

func (d *DocValuesLeafReader) GetNormValues(field string) (NumericDocValues, error) {
	return nil, errors.New("func GetNormValues is not yet implemented")
}

func (d *DocValuesLeafReader) GetFieldInfos() *FieldInfos {
	return nil
}

func (d *DocValuesLeafReader) GetLiveDocs() util.Bits {
	return nil
}

func (d *DocValuesLeafReader) GetPointValues(field string) (types.PointValues, bool) {
	return nil, false
}

func (d *DocValuesLeafReader) CheckIntegrity() error {
	return errors.New("func CheckIntegrity is not yet implemented")
}

func (d *DocValuesLeafReader) GetMetaData() *LeafMetaData {
	return nil
}
