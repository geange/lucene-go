package util

import "sort"

// Selector An implementation of a selection algorithm, ie. computing the k-th greatest value from a collection.
type Selector interface {
	// Select Reorder elements so that the element at position k is the same as if all elements were
	// sorted and all other elements are partitioned around it: [from, k) only contains elements that
	// are less than or equal to k and (k, to) only contains elements that are greater than or equal to k.
	Select(from, to, k int)

	// Swap values at slots i and j.
	Swap(i, j int)
}

func SelectorCheckArgs(from, to, k int) {
	if k < from {
		panic("k must be >= from")
	}
	if k >= to {
		panic("k must be < to")
	}
}

func SelectK(k int, data sort.Interface) {
	sort.Sort(data)
}
