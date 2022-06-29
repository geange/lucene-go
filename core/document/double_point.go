package document

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
