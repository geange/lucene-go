package document

type FloatRangeDocValuesField struct {
	*BinaryRangeDocValuesField

	field string
	min   []float32
	max   []float32
}
