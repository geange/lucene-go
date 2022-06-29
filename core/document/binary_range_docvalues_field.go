package document

type BinaryRangeDocValuesField struct {
	*BinaryDocValuesField

	field                string
	packedValue          []byte
	numDims              int
	numBytesPerDimension int
}

func NewBinaryRangeDocValuesField(field string,
	packedValue []byte, numDims int, numBytesPerDimension int) *BinaryRangeDocValuesField {
	return &BinaryRangeDocValuesField{
		BinaryDocValuesField: NewBinaryDocValuesField(field, packedValue),
		field:                field,
		packedValue:          packedValue,
		numDims:              numDims,
		numBytesPerDimension: numBytesPerDimension,
	}
}
