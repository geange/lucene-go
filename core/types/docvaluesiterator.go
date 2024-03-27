package types

type DocValuesIterator interface {
	DocIdSetIterator

	// AdvanceExact
	// Advance the iterator to exactly target and return whether target has a item.
	// target must be greater than or equal to the current doc ID and must be a valid doc ID, ie. ≥ 0 and < maxDoc.
	// After this method returns, docID() returns target.
	AdvanceExact(target int) (bool, error)
}
