package document

import (
	"bytes"
	"fmt"
	"github.com/geange/lucene-go/core/util/numeric"
)

// LongPoint
// An indexed long field for fast range filters. If you also need to store the value, you should add a separate StoredField instance.
// Finding all documents within an N-dimensional shape or range at search time is efficient. Multiple values for the same field in one document is allowed.
// This field defines static factory methods for creating common queries:
// newExactQuery(String, long) for matching an exact 1D point.
// newSetQuery(String, long...) for matching a set of 1D values.
// newRangeQuery(String, long, long) for matching a 1D range.
// newRangeQuery(String, long[], long[]) for matching points/ranges in n-dimensional space.
// See Also: PointValues
type LongPoint struct {
	*Field[[]byte]
}

// NewLongPoint
// Creates a new LongPoint, indexing the provided N-dimensional long point.
// Params: name – field name point – long[] value
// Throws: IllegalArgumentException – if the field name or value is null.
func NewLongPoint(name string, points ...int64) LongPoint {
	packed := packLongPoint(points)
	return LongPoint{NewField(name, packed, genLongPointType(len(points)))}
}

func (r *LongPoint) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("LongPoint")
	buf.WriteString(" <")
	buf.WriteString(r.name)
	buf.WriteString(":")

	packed := r.fieldsData
	count := r.fieldType.PointDimensionCount()
	for dim := 0; dim < count; dim++ {
		if dim > 0 {
			buf.WriteString(",")
		}
		offset := dim * LONG_BYTES
		num := fmt.Sprintf("%d", numeric.SortableBytesToLong(packed[offset:]))
		buf.WriteString(num)
	}
	buf.WriteString(">")
	return buf.String()
}

func (r *LongPoint) Number() (any, bool) {
	if r.fieldType.PointDimensionCount() > 1 {
		return int64(0), false
	}
	return numeric.SortableBytesToLong(r.fieldsData), true
}

func packLongPoint(points []int64) []byte {
	packed := make([]byte, len(points)*LONG_BYTES)
	for i, point := range points {
		offset := i * LONG_BYTES
		numeric.LongToSortableBytes(uint64(point), packed[offset:])
	}
	return packed
}

func genLongPointType(numDims int) *FieldType {
	fieldType := NewFieldType()
	_ = fieldType.SetDimensions(numDims, LONG_BYTES)
	fieldType.Freeze()
	return fieldType
}
