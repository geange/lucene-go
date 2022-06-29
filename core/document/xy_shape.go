package document

// XYShape A cartesian shape utility class for indexing and searching geometries whose vertices are
// unitless x, y values.
//This class defines seven static factory methods for common indexing and search operations:
//createIndexableFields(String, XYPolygon) for indexing a cartesian polygon.
//createIndexableFields(String, XYLine) for indexing a cartesian linestring.
//createIndexableFields(String, float, float) for indexing a x, y cartesian point.
//newBoxQuery() for matching cartesian shapes that have some ShapeField.QueryRelation with a bounding box.
//newLineQuery() for matching cartesian shapes that have some ShapeField.QueryRelation with a linestring.
//newPolygonQuery() for matching cartesian shapes that have some ShapeField.QueryRelation with a polygon.
//newGeometryQuery() for matching cartesian shapes that have some ShapeField.QueryRelation with one or more XYGeometry.
//WARNING: Like LatLonPoint, vertex values are indexed with some loss of precision from the original double values.
//See Also:
//PointValues, LatLonDocValuesField
type XYShape struct {
}
