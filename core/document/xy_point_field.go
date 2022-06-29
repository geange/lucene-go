package document

// XYPointField An indexed XY position field.
// Finding all documents within a range at search time is efficient. Multiple values for the same field in
// one document is allowed.
// This field defines static factory methods for common operations:
// newBoxQuery() for matching points within a bounding box.
// newDistanceQuery() for matching points within a specified distance.
// newPolygonQuery() for matching points within an arbitrary polygon.
// newGeometryQuery() for matching points within an arbitrary geometry collection.
// If you also need per-document operations such as sort by distance, add a separate XYDocValuesField instance.
// If you also need to store the value, you should add a separate StoredField instance.
// See Also: PointValues, XYDocValuesField
type XYPointField struct {
	*Field
}
