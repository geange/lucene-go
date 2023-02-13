package bkd

// PointValue Represents a dimensional point value written in the BKD tree.
// lucene.internal
type PointValue interface {

	// PackedValue Returns the packed values for the dimensions
	PackedValue() []byte

	// DocID Returns the docID
	DocID() int

	// PackedValueDocIDBytes Returns the byte representation of the packed value together with the docID
	PackedValueDocIDBytes() []byte
}
