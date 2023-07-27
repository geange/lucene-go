package index

// SortedSetDocValues A multi-valued version of SortedDocValues.
// Per-Document values in a SortedSetDocValues are deduplicated, dereferenced, and sorted into a
// dictionary of unique values. A pointer to the dictionary item (ordinal) can be retrieved for
// each document. Ordinals are dense and in increasing sorted order.
type SortedSetDocValues interface {
	DocValuesIterator

	NextOrd() (int64, error)

	LookupOrd(ord int64) ([]byte, error)

	GetValueCount() int64
}

const (
	NO_MORE_ORDS = -1
)
