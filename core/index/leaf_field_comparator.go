package index

import "github.com/geange/lucene-go/core/types"

// LeafFieldComparator Expert: comparator that gets instantiated on each leaf from a top-level FieldComparator instance.
// A leaf comparator must define these functions:
// setBottom This method is called by FieldValueHitQueue to notify the FieldComparator of the current weakest ("bottom") slot. Note that this slot may not hold the weakest item according to your comparator, in cases where your comparator is not the primary one (ie, is only used to break ties from the comparators before it).
// compareBottom Compare a new hit (docID) against the "weakest" (bottom) entry in the queue.
// compareTop Compare a new hit (docID) against the top item previously set by a call to FieldComparator.setTopValue.
// copy Installs a new hit into the priority queue. The FieldValueHitQueue calls this method when a new hit is competitive.
// See Also: FieldComparator
// lucene.experimental
type LeafFieldComparator interface {
	// SetBottom Set the bottom slot, ie the "weakest" (sorted last) entry in the queue. When compareBottom is called, you should compare against this slot. This will always be called before compareBottom.
	// Params: slot – the currently weakest (sorted last) slot in the queue
	SetBottom(slot int) error

	// CompareBottom compare the bottom of the queue with this doc. This will only invoked after setBottom has
	// been called. This should return the same result as FieldComparator.compare(int, int)} as if bottom were
	// slot1 and the new document were slot 2.
	// For a search that hits many results, this method will be the hotspot (invoked by far the most frequently).
	// Params: doc – that was hit
	// Returns: any N < 0 if the doc's item is sorted after the bottom entry (not competitive), any N > 0 if
	// the doc's item is sorted before the bottom entry and 0 if they are equal.
	CompareBottom(doc int) (int, error)

	// CompareTop compare the top item with this doc. This will only invoked after setTopValue has been called.
	// This should return the same result as FieldComparator.compare(int, int)} as if topValue were slot1 and
	// the new document were slot 2. This is only called for searches that use searchAfter (deep paging).
	// Params: doc – that was hit
	// Returns: any N < 0 if the doc's item is sorted after the top entry (not competitive), any N > 0 if the
	// doc's item is sorted before the top entry and 0 if they are equal.
	CompareTop(doc int) (int, error)

	// Copy This method is called when a new hit is competitive.
	// You should copy any state associated with this document that will be required for future comparisons,
	// into the specified slot.
	// Params:  slot – which slot to copy the hit to
	//			doc – docID relative to current reader
	Copy(slot, doc int) error

	// SetScorer Sets the Scorer to use in case a document's score is needed.
	// Params: scorer – Scorer instance that you should use to obtain the current hit's score, if necessary.
	SetScorer(scorer Scorable) error

	// CompetitiveIterator Returns a competitive iterator
	// Returns: an iterator over competitive docs that are stronger than already collected docs or
	// null if such an iterator is not available for the current comparator or segment.
	CompetitiveIterator() (types.DocIdSetIterator, error)

	// SetHitsThresholdReached Informs this leaf comparator that hits threshold is reached.
	// This method is called from a collector when hits threshold is reached.
	SetHitsThresholdReached() error
}
