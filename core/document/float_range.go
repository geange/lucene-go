package document

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
