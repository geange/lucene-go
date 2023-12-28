package document

import "sync"

var (
	latLonPointTypeOnce sync.Once
	latLonPointType     *FieldType
)

type LatLonPoint struct {
	*Field[*LatLon]
}

type LatLon struct {
	Latitude  float64
	Longitude float64
}

func NewLatLonPoint(name string, latitude, longitude float64) LatLonPoint {
	latLonPointTypeOnce.Do(func() {
		latLonPointType = NewFieldType()
		_ = latLonPointType.SetDimensions(2, INTEGER_BYTES)
	})

	value := &LatLon{
		Latitude:  latitude,
		Longitude: longitude,
	}
	return LatLonPoint{NewField(name, value, latLonPointType)}
}

// LatLonDocValuesField
// An per-document location field.
// Sorting by distance is efficient. Multiple values for the same field in one document is allowed.
// This field defines static factory methods for common operations:
// newDistanceSort() for ordering documents by distance from a specified location.
// If you also need query operations, you should add a separate LatLonPoint instance. If you also need to store
// the value, you should add a separate StoredField instance.
// WARNING: Values are indexed with some loss of precision from the original double values (4.190951585769653E-8
// for the latitude component and 8.381903171539307E-8 for longitude).
// See Also: LatLonPoint
type LatLonDocValuesField Field[LatLon]

type LatLonPointSortField struct {
}

type LatLonShape struct {
}
