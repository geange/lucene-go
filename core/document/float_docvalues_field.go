package document

import "math"

type FloatDocValuesField struct {
	*NumericDocValuesField
}

func NewFloatDocValuesField(name string, value float32) *FloatDocValuesField {
	bits := math.Float32bits(value)
	return &FloatDocValuesField{NewNumericDocValuesField(name, int(bits))}
}

func (r *FloatDocValuesField) SetFloat64(value float32) {
	r.Field.SetIntValue(int(math.Float32bits(value)))
}

func (r *FloatDocValuesField) SetIntValue(value int) {
	panic("")
}
