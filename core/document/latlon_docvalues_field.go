package document

// LatLonDocValuesField An per-document location field.
// Sorting by distance is efficient. Multiple values for the same field in one document is allowed.
// This field defines static factory methods for common operations:
// newDistanceSort() for ordering documents by distance from a specified location.
// If you also need query operations, you should add a separate LatLonPoint instance. If you also need to store
// the value, you should add a separate StoredField instance.
// WARNING: Values are indexed with some loss of precision from the original double values (4.190951585769653E-8
// for the latitude component and 8.381903171539307E-8 for longitude).
// See Also: LatLonPoint
type LatLonDocValuesField struct {
	*Field
}
