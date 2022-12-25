package index

// BinaryDocValues A per-document numeric value.
type BinaryDocValues interface {
	DocValuesIterator

	// BinaryValue Returns the binary value for the current document ID. It is illegal to call this method after
	// advanceExact(int) returned false.
	// Returns: binary value
	BinaryValue() ([]byte, error)
}
