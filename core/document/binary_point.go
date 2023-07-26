package document

import (
	"bytes"
	"errors"
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
	*Field
}

func NewBinaryPoint(name string, point ...[]byte) (*BinaryPoint, error) {
	packed, err := BinaryPointPack(point...)
	if err != nil {
		return nil, err
	}

	iType, err := BinaryPointGetType(point...)
	if err != nil {
		return nil, err
	}

	return NewBinaryPointWithType(name, packed, iType)
}

func NewBinaryPointWithType(name string, packedPoint []byte, iType IndexableFieldType) (*BinaryPoint, error) {
	return &BinaryPoint{NewField(name, packedPoint, iType)}, nil
}

func BinaryPointGetType(point ...[]byte) (*FieldType, error) {
	bytesPerDim := -1

	for i := 0; i < len(point); i++ {
		oneDim := point[i]
		if bytesPerDim == -1 {
			bytesPerDim = len(oneDim)
		} else if bytesPerDim != len(oneDim) {
			return nil, errors.New("all dimensions must have same bytes length")
		}
	}
	return binaryPointGetTypeV1(len(point), bytesPerDim)
}

func binaryPointGetTypeV1(numDims, bytesPerDim int) (*FieldType, error) {
	fType := NewFieldType()
	fType.SetDimensions(numDims, bytesPerDim)
	fType.Freeze()
	return fType, nil
}

func BinaryPointPack(point ...[]byte) ([]byte, error) {
	if len(point) == 1 {
		return point[0], nil
	}
	return bytes.Join(point, []byte{}), nil
}
