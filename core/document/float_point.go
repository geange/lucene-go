package document

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
