package index

const SINGLE_MEDIAN_THRESHOLD = 40

// IntroSorter Sorter implementation based on a variant of the quicksort algorithm called introsort :
// when the recursion level exceeds the log of the length of the array to sort, it falls back to heapsort.
// This prevents quicksort from running into its worst-case quadratic runtime. Small ranges are sorted with
// insertion sort.
// This sort algorithm is fast on most data shapes, especially with low cardinality. If the data to sort
// is known to be strictly ascending or descending, prefer TimSorter.
// lucene.internal
type IntroSorter struct {
	*Sort
}
