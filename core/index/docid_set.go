package index

import (
	"github.com/geange/lucene-go/core/util"
)

// A DocIdSet contains a set of doc ids. Implementing classes must only implement iterator to provide access to the set.
type DocIdSet interface {
	// DVFUIterator Provides a DocIdSetIterator to access the set.
	// This implementation can return null if there are no docs that match.
	Iterator() (DocIdSetIterator, error)

	// Bits
	// TODO: somehow this class should express the cost of
	// iteration vs the cost of random access Bits; for
	// expensive Filters (e.g. distance < 1 km) we should use
	// bits() after all other Query/Filters have matched, but
	// this is the opposite of what bits() is for now
	// (down-low filtering using e.g. FixedBitSet)
	// Optionally provides a Bits interface for random access to matching documents.
	// Returns: null, if this DocIdSet does not support random access. In contrast to iterator(),
	// a return value of null does not imply that no documents match the filter!
	// The default implementation does not provide random access,
	// so you only need to implement this method if your DocIdSet can guarantee random
	// access to every docid in O(1) time without external disk access
	// (as Bits interface cannot throw IOException). This is generally true for bit sets
	// like org.apache.lucene.util.FixedBitSet, which return itself if they are used as DocIdSet.
	Bits() (util.Bits, error)
}
