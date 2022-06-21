package core

// DocIdSetIterator This abstract class defines methods to iterate over a set of non-decreasing doc ids.
// Note that this class assumes it iterates on doc Ids, and therefore NO_MORE_DOCS is set to 2147483647
// in order to be used as a sentinel object. Implementations of this class are expected to consider
// Integer.MAX_VALUE as an invalid value.
type DocIdSetIterator interface {
	// DocID Returns the following:
	// * -1 if nextDoc() or advance(int) were not called yet.
	// * NO_MORE_DOCS if the iterator has exhausted.
	// * Otherwise it should return the doc ID it is currently on.
	// Since: 2.9
	DocID() int

	// NextDoc Advances to the next document in the set and returns the doc it is currently on, or
	// NO_MORE_DOCS if there are no more docs in the set. NOTE: after the iterator has exhausted
	// you should not call this method, as it may result in unpredicted behavior.
	// Since: 2.9
	NextDoc() (int, error)

	// Advance Advances to the first beyond the current whose document number is greater than or equal to
	// target, and returns the document number itself. Exhausts the iterator and returns NO_MORE_DOCS if
	// target is greater than the highest document number in the set.
	// The behavior of this method is undefined when called with target ≤ current, or after the iterator
	// has exhausted. Both cases may result in unpredicted behavior.
	// When target > current it behaves as if written:
	//     int advance(int target) {
	//       int doc;
	//       while ((doc = nextDoc()) < target) {
	//       }
	//       return doc;
	//     }
	//
	// Some implementations are considerably more efficient than that.
	// NOTE: this method may be called with NO_MORE_DOCS for efficiency by some Scorers. If your implementation
	// cannot efficiently determine that it should exhaust, it is recommended that you check for that value in
	// each call to this method.
	// Since: 2.9
	Advance(target int) (int, error)

	// SlowAdvance Slow (linear) implementation of advance relying on nextDoc() to advance beyond the target position.
	SlowAdvance(target int) (int, error)

	// Cost Returns the estimated cost of this DocIdSetIterator.
	// This is generally an upper bound of the number of documents this iterator might match, but may be a
	// rough heuristic, hardcoded value, or otherwise completely inaccurate.
	Cost() int64
}
