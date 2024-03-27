package search

type ScorerSupplier interface {
	// Get the Scorer. This may not return null and must be called at most once.
	// Params: 	leadCost – Cost of the scorer that will be used in order to lead iteration. This can be
	//			interpreted as an upper bound of the number of times that DocIdSetIterator.nextDoc,
	//			DocIdSetIterator.advance and TwoPhaseIterator.matches will be called. Under doubt,
	//			pass Long.MAX_VALUE, which will produce a Scorer that has good iteration capabilities.
	Get(leadCost int64) (Scorer, error)

	// Cost Get an estimate of the Scorer that would be returned by get. This may be a costly operation,
	// so it should only be called if necessary.
	// See Also: DocIdSetIterator.cost
	Cost() int64
}

type ScorerSupplierDefault struct {
}
