package util

type SortableBytesRefArray interface {

	// Append a new value
	Append(bytes *BytesRef) int

	// Clear all previously stored values
	Clear()

	// Size Returns the number of values appended so far
	Size() int

	// Iterator Sort all values by the provided comparator and return an iterator over the sorted values
	Iterator(comp *BytesRef) BytesRefIterator
}
