package document

import (
	"errors"
	"math"

	"github.com/geange/lucene-go/core/util/numeric"
)

type DoublePoint Float64Point

var NewDoublePoint = NewFloat64Point

// Float64Point
// An indexed double field for fast range filters. If you also need to store the value,
// you should add a separate StoredField instance.
// Finding all documents within an N-dimensional shape or range at search time is efficient.
// Multiple values for the same field in one document is allowed.
type Float64Point struct {
	*Field[[]byte]
}

func NewFloat64Point(name string, points ...float64) (*Float64Point, error) {
	if len(points) == 0 {
		return nil, errors.New("len(points) == 0")
	}

	packed := packFloat64Point(points)
	fieldType := genDoublePointType(len(points))
	return &Float64Point{NewField(name, packed, fieldType)}, nil
}

func (r *Float64Point) Number() (any, bool) {
	if r.fieldType.PointDimensionCount() > 1 {
		return float64(0), false
	}
	return decodeDimensionFloat64(r.fieldsData), true
}

func (r *Float64Point) Points() []float64 {
	return unPackFloat64Point(r.fieldsData)
}

func packFloat64Point(points []float64) []byte {
	packed := make([]byte, len(points)*DOUBLE_BYTES)
	for dim, point := range points {
		offset := dim * DOUBLE_BYTES
		encodeDimensionFloat64(point, packed[offset:])
	}
	return packed
}

func unPackFloat64Point(bs []byte) []float64 {
	points := make([]float64, 0, len(bs)/DOUBLE_BYTES)
	for i := 0; i < len(bs); i += DOUBLE_BYTES {
		point := decodeDimensionFloat64(bs[i:])
		points = append(points, point)
	}
	return points
}

func encodeDimensionFloat64(value float64, dest []byte) {
	numeric.Uint64ToSortableBytes(numeric.Float64ToSortableLong(value), dest)
}

func decodeDimensionFloat64(value []byte) float64 {
	return numeric.SortableUint64ToFloat64(numeric.SortableBytesToUint64(value))
}

func genDoublePointType(numDims int) *FieldType {
	fieldType := NewFieldType()
	_ = fieldType.SetDimensions(numDims, DOUBLE_BYTES)
	fieldType.Freeze()
	return fieldType
}

// DoubleDocValuesField
// Syntactic sugar for encoding doubles as NumericDocValues via Double.doubleToRawLongBits(double).
// Per-document double values can be retrieved via org.apache.lucene.index.LeafReader.getNumericDocValues(String).
// NOTE: In most all cases this will be rather inefficient, requiring eight bytes per document. Consider encoding
// double values yourself with only as much precision as you require.
type DoubleDocValuesField struct {
	NumericDocValuesField
}

func NewDoubleDocValuesField(name string, value float64) *DoubleDocValuesField {
	n := int64(math.Float64bits(value))
	return &DoubleDocValuesField{NewNumericDocValuesField(name, n)}
}
