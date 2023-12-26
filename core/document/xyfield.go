package document

import "math"

var (
	XYDocValuesFieldType *FieldType
)

func init() {
	XYDocValuesFieldType = NewFieldType()
	_ = XYDocValuesFieldType.SetDocValuesType(DOC_VALUES_TYPE_SORTED_NUMERIC)
	XYDocValuesFieldType.Freeze()
}

// XYDocValuesField An per-document location field.
// Sorting by distance is efficient. Multiple values for the same field in one document is allowed.
// This field defines static factory methods for common operations:
// newSlowBoxQuery() for matching points within a bounding box.
// newSlowDistanceQuery() for matching points within a specified distance.
// newSlowPolygonQuery() for matching points within an arbitrary polygon.
// newSlowGeometryQuery() for matching points within an arbitrary geometry.
// newDistanceSort() for ordering documents by distance from a specified location.
// If you also need query operations, you should add a separate XYPointField instance. If you also need to
// store the value, you should add a separate StoredField instance.
// See Also: XYPointField
type XYDocValuesField struct {
	*Field
}

func NewXYDocValuesField(name string, x, y float32) *XYDocValuesField {
	field := &XYDocValuesField{NewFieldV1(name, XYDocValuesFieldType)}
	field.setLocationValue(x, y)
	return field
}

func (r *XYDocValuesField) setLocationValue(x, y float32) {
	xEncoded := int64(math.Float32bits(x))
	yEncoded := int64(math.Float32bits(y))
	r.fieldsData = xEncoded<<32 | yEncoded
}

// XYShape A cartesian shape utility class for indexing and searching geometries whose vertices are
// unitless x, y values.
// This class defines seven static factory methods for common indexing and search operations:
// createIndexableFields(String, XYPolygon) for indexing a cartesian polygon.
// createIndexableFields(String, XYLine) for indexing a cartesian linestring.
// createIndexableFields(String, float, float) for indexing a x, y cartesian point.
// newBoxQuery() for matching cartesian shapes that have some ShapeField.QueryRelation with a bounding box.
// newLineQuery() for matching cartesian shapes that have some ShapeField.QueryRelation with a linestring.
// newPolygonQuery() for matching cartesian shapes that have some ShapeField.QueryRelation with a polygon.
// newGeometryQuery() for matching cartesian shapes that have some ShapeField.QueryRelation with one or more XYGeometry.
// WARNING: Like LatLonPoint, vertex values are indexed with some loss of precision from the original double values.
// See Also:
// PointValues, LatLonDocValuesField
type XYShape struct {
}

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

type XYPointSortField struct {
	*Field
}
