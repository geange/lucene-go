package index

const (
	BINARY_SORT_THRESHOLD    = 20
	INSERTION_SORT_THRESHOLD = 16
)

// Sorter Base class for sorting algorithms implementations.
// lucene.internal
type Sorter interface {
	// Compare entries found in slots i and j. The contract for the returned value is the same as Comparator.compare(Object, Object).
	Compare(i, j int) int

	Swap(i, j int) int
}

type SorterDefault struct {
	pivotIndex int
	fnCompare  func(i, j int) int
	fnSwap     func(i, j int)
}
