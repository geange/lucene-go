package index

// SortedNumericDocValues A list of per-document numeric values, sorted according to Long.CompareFn(long, long).
type SortedNumericDocValues interface {
	DocValuesIterator

	// NextValue Iterates to the next item in the current document. Do not call this more than
	// docValueCount times for the document.
	NextValue() (int64, error)

	// DocValueCount Retrieves the number of values for the current document. This must always be greater
	// than zero. It is illegal to call this method after advanceExact(int) returned false.
	DocValueCount() int
}
