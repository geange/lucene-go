package document

import "math"

// FloatPoint
// An indexed float field for fast range filters. If you also need to store the value, you should
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
	*Field[[]float32]
}

func NewFloatPoint(name string, points ...float32) (FloatPoint, error) {
	fieldType := NewFieldType()
	if err := fieldType.SetDimensions(len(points), FLOAT_BYTES); err != nil {
		return FloatPoint{}, err
	}
	fieldType.Freeze()

	field := FloatPoint{NewField(name, points, fieldType)}
	return field, nil
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
	*Field[[]byte]
}

type FloatRangeDocValuesField struct {
	*BinaryRangeDocValuesField

	field string
	min   []float32
	max   []float32
}

type FloatDocValuesField struct {
	NumericDocValuesField
}

// NewFloatDocValuesField
// Syntactic sugar for encoding floats as NumericDocValues via Float.floatToRawIntBits(float).
// Per-document floating point values can be retrieved via org.apache.lucene.index.LeafReader.getNumericDocValues(String).
// NOTE: In most all cases this will be rather inefficient, requiring four bytes per document. Consider encoding floating point values yourself with only as much precision as you require.
func NewFloatDocValuesField(name string, value float32) *FloatDocValuesField {
	return &FloatDocValuesField{
		NewNumericDocValuesField(name, int64(math.Float32bits(value))),
	}
}
