package index

import (
	"errors"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

var _ index.LeafReader = &DocValuesLeafReader{}

type DocValuesLeafReader struct {
	*BaseLeafReader
}

func NewDocValuesLeafReader() *DocValuesLeafReader {
	reader := &DocValuesLeafReader{}

	reader.BaseLeafReader = NewBaseLeafReader(reader)
	return reader
}

func (d *DocValuesLeafReader) GetTermVectors(docID int) (index.Fields, error) {
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

func (d *DocValuesLeafReader) GetReaderCacheHelper() index.CacheHelper {
	return nil
}

func (d *DocValuesLeafReader) Terms(field string) (index.Terms, error) {
	return nil, errors.New("func Terms is not yet implemented")
}

func (d *DocValuesLeafReader) GetNumericDocValues(field string) (index.NumericDocValues, error) {
	return nil, errors.New("func NumericDocValues is not yet implemented")
}

func (d *DocValuesLeafReader) GetBinaryDocValues(field string) (index.BinaryDocValues, error) {
	return nil, errors.New("func BinaryDocValues is not yet implemented")
}

func (d *DocValuesLeafReader) GetSortedDocValues(field string) (index.SortedDocValues, error) {
	return nil, errors.New("func SortedDocValues is not yet implemented")
}

func (d *DocValuesLeafReader) GetSortedNumericDocValues(field string) (index.SortedNumericDocValues, error) {
	return nil, errors.New("func SortedNumericDocValues is not yet implemented")
}

func (d *DocValuesLeafReader) GetSortedSetDocValues(field string) (index.SortedSetDocValues, error) {
	return nil, errors.New("func SortedSetDocValues is not yet implemented")
}

func (d *DocValuesLeafReader) GetNormValues(field string) (index.NumericDocValues, error) {
	return nil, errors.New("func GetNormValues is not yet implemented")
}

func (d *DocValuesLeafReader) GetFieldInfos() index.FieldInfos {
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

func (d *DocValuesLeafReader) GetMetaData() index.LeafMetaData {
	return nil
}
