package document

// BinaryPoint
// An indexed binary field for fast range filters. If you also need to store the value,
// you should add a separate StoredField instance.
// Finding all documents within an N-dimensional shape or range at search time is efficient.
// Multiple values for the same field in one document is allowed.
// This field defines static factory methods for creating common queries:
// * NewExactQuery(String, byte[]) for matching an exact 1D point.
// * newSetQuery(String, byte[]...) for matching a set of 1D values.
// * newRangeQuery(String, byte[], byte[]) for matching a 1D range.
// * newRangeQuery(String, byte[][], byte[][]) for matching points/ranges in n-dimensional space.
type BinaryPoint struct {
	*Field
}

func NewBinaryPoint(name string, point ...[]byte) (*BinaryPoint, error) {
	packed, err := BinaryPointPack(point...)
	if err != nil {
		return nil, err
	}

	iType, err := BinaryPointGetType(point...)
	if err != nil {
		return nil, err
	}

	return NewBinaryPointWithType(name, packed, iType)
}

func NewBinaryPointWithType(name string, packedPoint []byte, iType IndexableFieldType) (*BinaryPoint, error) {
	panic("")
}

func BinaryPointGetType(point ...[]byte) (*FieldType, error) {
	panic("")
}

func BinaryPointPack(point ...[]byte) ([]byte, error) {
	panic("")
}
