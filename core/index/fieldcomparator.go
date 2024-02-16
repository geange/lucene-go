package index

// FieldComparator Expert: a FieldComparator compares hits so as to determine their sort order when collecting the top results with TopFieldCollector. The concrete public FieldComparator classes here correspond to the SortField types.
// The document IDs passed to these methods must only move forwards, since they are using doc values iterators to retrieve sort values.
// This API is designed to achieve high performance sorting, by exposing a tight interaction with FieldValueHitQueue as it visits hits. Whenever a hit is competitive, it's enrolled into a virtual slot, which is an int ranging from 0 to numHits-1. Segment transitions are handled by creating a dedicated per-segment LeafFieldComparator which also needs to interact with the FieldValueHitQueue but can optimize based on the segment to collect.
// The following functions need to be implemented
// compare Compare a hit at 'slot a' with hit 'slot b'.
// setTopValue This method is called by TopFieldCollector to notify the FieldComparator of the top most item, which is used by future calls to LeafFieldComparator.compareTop.
// getLeafComparator(LeafReaderContext) Invoked when the search is switching to the next segment. You may need to update internal state of the comparator, for example retrieving new values from DocValues.
// item Return the sort item stored in the specified slot. This is only called at the end of the search, in order to populate FieldDoc.fields when returning the top results.
// See Also: LeafFieldComparator
// lucene.experimental
type FieldComparator interface {
	// Compare hit at slot1 with hit at slot2.
	// Params: 	slot1 – first slot to compare
	//			slot2 – second slot to compare
	// Returns: any N < 0 if slot2's item is sorted after slot1, any N > 0 if the slot2's item is sorted
	// before slot1 and 0 if they are equal
	Compare(slot1, slot2 int) int

	// SetTopValue Record the top item, for future calls to LeafFieldComparator.compareTop. This is only
	// called for searches that use searchAfter (deep paging), and is called before any calls to
	// getLeafComparator(LeafReaderContext).
	SetTopValue(value any)

	// Value Return the actual item in the slot.
	// Params: slot – the item
	// Returns: item in this slot
	Value(slot int) any

	// GetLeafComparator Get a per-segment LeafFieldComparator to collect the given LeafReaderContext.
	// All docIDs supplied to this LeafFieldComparator are relative to the current reader (you must
	// add docBase if you need to map it to a top-level docID).
	// Params: context – current reader context
	// Returns: the comparator to use for this segment
	// Throws: IOException – if there is a low-level IO error
	GetLeafComparator(context *LeafReaderContext) (LeafFieldComparator, error)

	// CompareValues
	// Returns a negative integer if first is less than second,
	// 0 if they are equal and a positive integer otherwise.
	// Default impl to assume the type implements Comparable and invoke .compareTo;
	// be sure to override this method if your FieldComparator's type isn't a
	// Comparable or if your values may sometimes be null
	CompareValues(first, second any) int

	// SetSingleSort Informs the comparator that sort is done on this single field.
	// This is useful to enable some optimizations for skipping non-competitive documents.
	SetSingleSort()

	// DisableSkipping Informs the comparator that the skipping of documents should be disabled.
	// This function is called in cases when the skipping functionality should not be applied
	// or not necessary. One example for numeric comparators is when we don't know if the same
	// numeric data has been indexed with docValues and points if these two fields have the
	// same name. As the skipping functionality relies on these fields to have the same data
	// and as we don't know if it is true, we have to disable it. Another example could be
	// when search sort is a part of the index sort, and can be already efficiently handled by
	// TopFieldCollector, and doing extra work for skipping in the comparator is redundant.
	DisableSkipping()
}

// FieldComparatorSource Provides a FieldComparator for custom field sorting.
// lucene.experimental
type FieldComparatorSource interface {
	NewComparator(fieldName string, numHits, sortPos int, reversed bool) FieldComparator
}
