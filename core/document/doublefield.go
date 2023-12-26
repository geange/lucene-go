package document

import "math"

// DoublePoint An indexed double field for fast range filters. If you also need to store the value,
// you should add a separate StoredField instance.
// Finding all documents within an N-dimensional shape or range at search time is efficient.
// Multiple values for the same field in one document is allowed.
// This field defines static factory methods for creating common queries:
// newExactQuery(String, double) for matching an exact 1D point.
// newSetQuery(String, double...) for matching a set of 1D values.
// newRangeQuery(String, double, double) for matching a 1D range.
// newRangeQuery(String, double[], double[]) for matching points/ranges in n-dimensional space.
// See Also: PointValues
type DoublePoint struct {
	*Field
}

// DoubleDocValuesField Syntactic sugar for encoding doubles as NumericDocValues via Double.doubleToRawLongBits(double).
// Per-document double values can be retrieved via org.apache.lucene.index.LeafReader.getNumericDocValues(String).
// NOTE: In most all cases this will be rather inefficient, requiring eight bytes per document. Consider encoding
// double values yourself with only as much precision as you require.
type DoubleDocValuesField struct {
	*NumericDocValuesField
}

func NewDoubleDocValuesField(name string, value float64) *DoubleDocValuesField {
	bits := math.Float64bits(value)
	return &DoubleDocValuesField{NewNumericDocValuesField(name, int(bits))}
}

func (r *DoubleDocValuesField) SetFloat64(value float64) {
	r.Field.SetIntValue(int(math.Float64bits(value)))
}

func (r *DoubleDocValuesField) SetIntValue(value int) {
	panic("")
}
