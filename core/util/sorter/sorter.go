package sorter

type Sorter interface {
	Sort(from, to int)
}

const (
	// INSERTION_SORT_THRESHOLD
	// Below this size threshold, the sub-range is sorted using Insertion sort.
	INSERTION_SORT_THRESHOLD = 16
)
