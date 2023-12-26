package document

import (
	"math"

	"github.com/geange/lucene-go/core/util/numeric"
)

// DoublePoint
// An indexed double field for fast range filters. If you also need to store the value,
// you should add a separate StoredField instance.
// Finding all documents within an N-dimensional shape or range at search time is efficient.
// Multiple values for the same field in one document is allowed.
// See Also: PointValues
type DoublePoint struct {
	*Field[[]byte]
}

func NewDoublePoint(name string, points ...float64) *DoublePoint {
	packed := packDoublePoint(points)
	fieldType := genDoublePointType(len(points))
	return &DoublePoint{NewField(name, packed, fieldType)}
}

func (r *DoublePoint) Number() (any, bool) {
	if r.fieldType.PointDimensionCount() > 1 {
		return float64(0), false
	}
	return decodeDimensionFloat64(r.fieldsData), true
}

func packDoublePoint(points []float64) []byte {
	packed := make([]byte, len(points)*DOUBLE_BYTES)
	for dim, point := range points {
		offset := dim * DOUBLE_BYTES
		encodeDimensionFloat64(point, packed[offset:])
	}
	return packed
}

func encodeDimensionFloat64(value float64, dest []byte) {
	numeric.LongToSortableBytes(numeric.DoubleToSortableLong(value), dest)
}

func decodeDimensionFloat64(value []byte) float64 {
	return numeric.SortableLongToDouble(numeric.SortableBytesToUint64(value))
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
