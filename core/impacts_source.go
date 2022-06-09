package core

// ImpactsSource Source of Impacts.
type ImpactsSource interface {
	// AdvanceShallow Shallow-advance to target. This is cheaper than calling DocIdSetIterator.advance(int) and allows further calls to getImpacts() to ignore doc IDs that are less than target in order to get more precise information about impacts. This method may not be called on targets that are less than the current DocIdSetIterator.docID(). After this method has been called, DocIdSetIterator.nextDoc() may not be called if the current doc ID is less than target - 1 and DocIdSetIterator.advance(int) may not be called on targets that are less than target.
	AdvanceShallow(target int) error

	// GetImpacts Get information about upcoming impacts for doc ids that are greater than or equal to the maximum of DocIdSetIterator.docID() and the last target that was passed to advanceShallow(int). This method may not be called on an unpositioned iterator on which advanceShallow(int) has never been called. NOTE: advancing this iterator may invalidate the returned impacts, so they should not be used after the iterator has been advanced.
	GetImpacts() (Impacts, error)
}
