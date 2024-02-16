package document

import (
	"bytes"
	"fmt"

	"github.com/geange/lucene-go/core/util/numeric"
)

type IntPoint struct {
	*Field[[]byte]
}

func NewIntPoint(name string, points ...int32) IntPoint {
	packed := packIntPoint(points)
	fieldType := genIntPointType(len(points))
	return IntPoint{NewField(name, packed, fieldType)}
}

func (r *IntPoint) String() string {
	result := new(bytes.Buffer)
	result.WriteString("IntPoint")
	result.WriteString(" <")
	result.WriteString(r.name)
	result.WriteString(":")

	count := r.fieldType.PointDimensionCount()
	for dim := 0; dim < count; dim++ {
		if dim > 0 {
			result.WriteString(",")
		}
		offset := dim * INTEGER_BYTES
		num := fmt.Sprintf("%d", decodeDimensionInt32(r.fieldsData[offset:]))
		result.WriteString(num)
	}
	result.WriteString(">")
	return result.String()
}

func (r *IntPoint) Number() (any, bool) {
	if r.fieldType.PointDimensionCount() > 1 {
		return int32(0), false
	}
	return decodeDimensionInt32(r.fieldsData), true
}

func packIntPoint(points []int32) []byte {
	packed := make([]byte, len(points)*INTEGER_BYTES)

	for i, point := range points {
		offset := i * INTEGER_BYTES
		encodeDimensionInt32(point, packed[offset:])
	}
	return packed
}

func genIntPointType(numDims int) *FieldType {
	fieldType := NewFieldType()
	_ = fieldType.SetDimensions(numDims, INTEGER_BYTES)
	fieldType.Freeze()
	return fieldType
}

func encodeDimensionInt32(value int32, dest []byte) {
	numeric.IntToSortableBytes(value, dest)
}

func decodeDimensionInt32(value []byte) int32 {
	return numeric.SortableBytesToInt(value)
}
