package document

import "math"

type FloatDocValuesField struct {
	*NumericDocValuesField
}

func NewFloatDocValuesField(name string, value float32) *FloatDocValuesField {
	bits := math.Float32bits(value)
	return &FloatDocValuesField{NewNumericDocValuesField(name, int(bits))}
}

func (r *FloatDocValuesField) SetFloat64(value float32) {
	r.Field.SetIntValue(int(math.Float32bits(value)))
}

func (r *FloatDocValuesField) SetIntValue(value int) {
	panic("")
}

// FloatPoint An indexed float field for fast range filters. If you also need to store the value, you should
// add a separate StoredField instance.
// Finding all documents within an N-dimensional at search time is efficient. Multiple values for the same
// field in one document is allowed.
// This field defines static factory methods for creating common queries:
// newExactQuery(String, float) for matching an exact 1D point.
// newSetQuery(String, float...) for matching a set of 1D values.
// newRangeQuery(String, float, float) for matching a 1D range.
// newRangeQuery(String, float[], float[]) for matching points/ranges in n-dimensional space.
// See Also: PointValues
type FloatPoint struct {
	*Field
}

// FloatRange An indexed Float Range field.
// This field indexes dimensional ranges defined as min/max pairs. It supports up to a maximum of 4 dimensions
// (indexed as 8 numeric values). With 1 dimension representing a single float range, 2 dimensions representing
// a bounding box, 3 dimensions a bounding cube, and 4 dimensions a tesseract.
// Multiple values for the same field in one document is supported, and open ended ranges can be defined using
// Float.NEGATIVE_INFINITY and Float.POSITIVE_INFINITY.
// This field defines the following static factory methods for common search operations over float ranges:
// newIntersectsQuery() matches ranges that intersect the defined search range.
// newWithinQuery() matches ranges that are within the defined search range.
// newContainsQuery() matches ranges that contain the defined search range.
type FloatRange struct {
	*Field
}

type FloatRangeDocValuesField struct {
	*BinaryRangeDocValuesField

	field string
	min   []float32
	max   []float32
}
