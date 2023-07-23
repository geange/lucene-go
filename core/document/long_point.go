package document

import (
	"errors"
	"fmt"
	"github.com/geange/lucene-go/core/util/numeric"
)

const (
	LONG_BYTES = 8
)

// LongPoint
// An indexed long field for fast range filters. If you also need to store the value, you should add a separate StoredField instance.
// Finding all documents within an N-dimensional shape or range at search time is efficient. Multiple values for the same field in one document is allowed.
// This field defines static factory methods for creating common queries:
// newExactQuery(String, long) for matching an exact 1D point.
// newSetQuery(String, long...) for matching a set of 1D values.
// newRangeQuery(String, long, long) for matching a 1D range.
// newRangeQuery(String, long[], long[]) for matching points/ranges in n-dimensional space.
// See Also: PointValues
type LongPoint struct {
	*Field
}

// NewLongPoint
// Creates a new LongPoint, indexing the provided N-dimensional long point.
// Params: name – field name point – long[] value
// Throws: IllegalArgumentException – if the field name or value is null.
func NewLongPoint(name string, point ...int64) (*LongPoint, error) {
	packed, err := packLongs(point...)
	if err != nil {
		return nil, err
	}
	return &LongPoint{
		NewField(name, packed, LongPointGetType(len(point))),
	}, nil
}

func (r *LongPoint) SetLongValue(value int64) error {
	return r.SetLongValues(value)
}

// SetLongValues
// Change the values of this field
func (r *LongPoint) SetLongValues(points ...int64) error {
	if r._type.PointIndexDimensionCount() != len(points) {
		format := "this field(%s) uses %d dimensions; cannot change to (incoming) %d dimensions"
		return fmt.Errorf(format, r.name, r._type.PointIndexDimensionCount(), len(points))
	}

	packed, err := packLongs(points...)
	if err != nil {
		return err
	}
	r.fieldsData = packed
	return nil
}

func (r *LongPoint) SetBytesValue(bs []byte) error {
	return errors.New("cannot change value type from int64 to bytes")
}

func (r *LongPoint) NumericValue() {

}

// Pack a long point into a BytesRef
// point: long[] value
// Throws: IllegalArgumentException – is the value is null or of zero length
func packLongs(points ...int64) ([]byte, error) {
	if len(points) == 0 {
		return nil, errors.New("points must not be 0 dimensions")
	}

	packed := make([]byte, len(points)*LONG_BYTES)

	for dim := 0; dim < len(points); dim++ {
		encodeDimension(points[dim], packed[dim*LONG_BYTES:])
	}
	return packed, nil
}

func LongPointGetType(numDims int) *FieldType {
	fieldType := NewFieldType()
	_ = fieldType.SetDimensions(numDims, LONG_BYTES)
	fieldType.Freeze()
	return fieldType
}

func encodeDimension(value int64, dest []byte) {
	numeric.LongToSortableBytes(value, dest)
}

// Decode single long dimension
func decodeDimension(value []byte) int64 {
	return numeric.SortableBytesToLong(value)
}
