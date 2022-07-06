package index

type NumericDocValues interface {
	DocValuesIterator

	// LongValue Returns the numeric value for the current document ID. It is illegal to call this method
	// after advanceExact(int) returned false.
	// Returns: numeric value
	LongValue() (int64, error)
}
