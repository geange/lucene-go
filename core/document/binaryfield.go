package document

import (
	"bytes"
	"errors"
	"sync"
)

// BinaryPoint
// An indexed binary field for fast range filters. If you also need to store the value,
// you should add a separate StoredField instance.
// Finding all documents within an N-dimensional shape or range at search time is efficient.
// Multiple values for the same field in one document is allowed.
// This field defines static factory methods for creating common queries:
// * NewExactQuery(String, byte[]) for matching an exact 1D point.
// * newSetQuery(String, byte[]...) for matching a set of 1D values.
// * newRangeQuery(String, byte[], byte[]) for matching a 1D range.
// * newRangeQuery(String, byte[][], byte[][]) for matching points/ranges in n-dimensional space.
type BinaryPoint struct {
	*Field[[]byte]
}

func NewBinaryPoint(name string, points ...[]byte) (*BinaryPoint, error) {
	fieldType, err := genBinaryPointType(points)
	if err != nil {
		return nil, err
	}

	packed := bytes.Join(points, []byte{})

	return &BinaryPoint{NewField(name, packed, fieldType)}, nil
}

func genBinaryPointType(point [][]byte) (*FieldType, error) {
	bytesPerDim := -1

	for i := 0; i < len(point); i++ {
		oneDim := point[i]
		if bytesPerDim == -1 {
			bytesPerDim = len(oneDim)
		} else if bytesPerDim != len(oneDim) {
			return nil, errors.New("all dimensions must have same bytes length")
		}
	}

	fType := NewFieldType()
	if err := fType.SetDimensions(len(point), bytesPerDim); err != nil {
		return nil, err
	}
	fType.Freeze()
	return fType, nil
}

var (
	binaryDocValuesFieldTypeOnce sync.Once
	binaryDocValuesFieldType     *FieldType
)

// BinaryDocValuesField
// Field that stores a per-document value([]byte).
// The values are stored directly with no sharing, which is a good fit when the fields don't share (many)
// values, such as a title field. If values may be shared and sorted it's better to use SortedDocValuesField.
// If you also need to store the value, you should add a separate StoredField instance.
// See Also: index.BinaryDocValues
type BinaryDocValuesField struct {
	*Field[[]byte]
}

func NewBinaryDocValuesField(name string, value []byte) *BinaryDocValuesField {
	binaryDocValuesFieldTypeOnce.Do(func() {
		binaryDocValuesFieldType = NewFieldType()
		_ = binaryDocValuesFieldType.SetDocValuesType(DOC_VALUES_TYPE_BINARY)
		binaryDocValuesFieldType.Freeze()
	})

	return &BinaryDocValuesField{NewField(name, value, binaryDocValuesFieldType)}
}

type BinaryRangeDocValuesField struct {
	*BinaryDocValuesField

	field                string
	packedValue          []byte
	numDims              int
	numBytesPerDimension int
}

func NewBinaryRangeDocValuesField(field string,
	packedValue []byte, numDims int, numBytesPerDimension int) *BinaryRangeDocValuesField {

	valuesField := NewBinaryDocValuesField(field, packedValue)
	return &BinaryRangeDocValuesField{
		BinaryDocValuesField: valuesField,
		field:                field,
		packedValue:          packedValue,
		numDims:              numDims,
		numBytesPerDimension: numBytesPerDimension,
	}
}
