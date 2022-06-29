package document

import (
	"math"
)

// DoubleDocValuesField Syntactic sugar for encoding doubles as NumericDocValues via Double.doubleToRawLongBits(double).
// Per-document double values can be retrieved via org.apache.lucene.index.LeafReader.getNumericDocValues(String).
// NOTE: In most all cases this will be rather inefficient, requiring eight bytes per document. Consider encoding
// double values yourself with only as much precision as you require.
type DoubleDocValuesField struct {
	*NumericDocValuesField
}

func NewDoubleDocValuesField(name string, value float64) *DoubleDocValuesField {
	bits := math.Float64bits(value)
	return &DoubleDocValuesField{NewNumericDocValuesField(name, int(bits))}
}

func (r *DoubleDocValuesField) SetFloat64(value float64) {
	r.Field.SetIntValue(int(math.Float64bits(value)))
}

func (r *DoubleDocValuesField) SetIntValue(value int) {
	panic("")
}
