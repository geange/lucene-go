package index

// Sorter Base class for sorting algorithms implementations.
// lucene.internal
type Sorter interface {
	// Compare entries found in slots i and j.
	// The contract for the returned item is the same as cmp.CompareFn(Object, Object).
	Compare(i, j int) int

	Swap(i, j int) int
}
