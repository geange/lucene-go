package document

import (
	"github.com/geange/lucene-go/core/types"
	"math"
)

var (
	XYDocValuesFieldType *FieldType
)

func init() {
	XYDocValuesFieldType = NewFieldType()
	XYDocValuesFieldType.SetDocValuesType(types.DOC_VALUES_TYPE_SORTED_NUMERIC)
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
