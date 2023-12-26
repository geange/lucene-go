package packed

// Reader
// A read-only random access array of positive integers.
// lucene.internal
type Reader interface {
	// Get the long at the given index. Behavior is undefined for out-of-range indices.
	Get(index int) uint64

	// GetBulk Bulk get: read at least one and at most len longs starting from index into
	// arr[off:off+len] and return the actual number of values that have been read.
	GetBulk(index int, arr []uint64) int

	// Size Returns: the number of values.
	Size() int
}
