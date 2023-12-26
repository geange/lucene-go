package document

import (
	"bytes"
	"errors"
)

var (
	binaryDocValuesFieldType *FieldType
)

func init() {
	binaryDocValuesFieldType = NewFieldType()
	_ = binaryDocValuesFieldType.SetDocValuesType(DOC_VALUES_TYPE_BINARY)
	binaryDocValuesFieldType.Freeze()
}

// BinaryDocValuesField Field that stores a per-document value([]byte).
// The values are stored directly with no sharing, which is a good fit when the fields don't share (many)
// values, such as a title field. If values may be shared and sorted it's better to use SortedDocValuesField.
// Here's an example usage:
//
//	document.add(new BinaryDocValuesField(name, new BytesRef("hello")));
//
// If you also need to store the value, you should add a separate StoredField instance.
// See Also:
// BinaryDocValues
type BinaryDocValuesField struct {
	*Field
}

func NewBinaryDocValuesField(name string, value []byte) *BinaryDocValuesField {
	field := &BinaryDocValuesField{NewFieldV1(name, binaryDocValuesFieldType)}
	field.fieldsData = value
	return field
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
	return &BinaryRangeDocValuesField{
		BinaryDocValuesField: NewBinaryDocValuesField(field, packedValue),
		field:                field,
		packedValue:          packedValue,
		numDims:              numDims,
		numBytesPerDimension: numBytesPerDimension,
	}
}

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
	*Field
}

func NewBinaryPoint(name string, point ...[]byte) (*BinaryPoint, error) {
	packed, err := BinaryPointPack(point...)
	if err != nil {
		return nil, err
	}

	iType, err := binaryPointGetType(point...)
	if err != nil {
		return nil, err
	}

	return &BinaryPoint{NewField(name, packed, iType)}, nil
}

func binaryPointGetType(point ...[]byte) (*FieldType, error) {
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

func BinaryPointPack(point ...[]byte) ([]byte, error) {
	if len(point) == 1 {
		return point[0], nil
	}
	return bytes.Join(point, []byte{}), nil
}
