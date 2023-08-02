package search

// DisiPriorityQueue
// A priority queue of DocIdSetIterators that orders by current doc ID. This specialization is needed over PriorityQueue because the pluggable comparison function makes the rebalancing quite slow.
// lucene.internal
type DisiPriorityQueue struct {
}
