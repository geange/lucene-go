package types

import (
	"context"
	"io"
	"math"
)

// DocIdSetIterator
// This abstract class defines methods to iterate over a set of non-decreasing doc ids.
// Note that this class assumes it iterates on doc Ids, and therefore NO_MORE_DOCS is set to 2147483647
// in order to be used as a sentinel object. Implementations of this class are expected to consider
// Integer.MAX_VALUE as an invalid item.
type DocIdSetIterator interface {
	// DocID
	// Returns the following:
	// * -1 if nextDoc() or advance(int) were not called yet.
	// * NO_MORE_DOCS if the iterator has exhausted.
	// * Otherwise it should return the doc ID it is currently on.
	// Since: 2.9
	DocID() int

	// NextDoc
	// Advances to the next document in the set and returns the doc it is currently on, or
	// NO_MORE_DOCS if there are no more docs in the set. NOTE: after the iterator has exhausted
	// you should not call this method, as it may result in unpredicted behavior.
	// Since: 2.9
	NextDoc(ctx context.Context) (int, error)

	// Advance
	// Advances to the first beyond the current whose document number is greater than or equal to
	// target, and returns the document number itself. Exhausts the iterator and returns NO_MORE_DOCS if
	// target is greater than the highest document number in the set.
	// The behavior of this method is undefined when called with target ≤ current, or after the iterator
	// has exhausted. Both cases may result in unpredicted behavior.
	// When target > current it behaves as if written:
	//     func advance(target int) int {
	//       var doc int
	//       for ((doc = nextDoc()) < target) {
	//       }
	//       return doc;
	//     }
	//
	// Some implementations are considerably more efficient than that.
	// NOTE: this method may be called with NO_MORE_DOCS for efficiency by some Scorers. If your implementation
	// cannot efficiently determine that it should exhaust, it is recommended that you check for that item in
	// each call to this method.
	// Since: 2.9
	Advance(ctx context.Context, target int) (int, error)

	// SlowAdvance
	// Slow (linear) implementation of advance relying on nextDoc() to advance beyond the target position.
	SlowAdvance(ctx context.Context, target int) (int, error)

	// Cost
	// Returns the estimated cost of this DocIdSetIterator.
	// This is generally an upper bound of the number of documents this iterator might match, but may be a
	// rough heuristic, hardcoded item, or otherwise completely inaccurate.
	Cost() int64
}

const (
	NO_MORE_DOCS = math.MaxInt32
)

type DocNext interface {
	NextDoc(ctx context.Context) (int, error)
}

func SlowAdvanceWithContext(ctx context.Context, next DocNext, target int) (int, error) {
	doc := 0

	var err error
	for doc < target {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			doc, err = next.NextDoc(ctx)
			if err != nil {
				return 0, err
			}
		}
	}
	return doc, nil
}

func DocIdSetIteratorAll(maxDoc int) DocIdSetIterator {
	return &docIdSetIteratorAll{
		doc:    -1,
		maxDoc: maxDoc,
	}
}

var _ DocIdSetIterator = &docIdSetIteratorAll{}

type docIdSetIteratorAll struct {
	doc    int
	maxDoc int
}

func (d *docIdSetIteratorAll) SlowAdvance(ctx context.Context, target int) (int, error) {
	return SlowAdvanceWithContext(ctx, d, target)
}

func (d *docIdSetIteratorAll) DocID() int {
	return d.doc
}

func (d *docIdSetIteratorAll) NextDoc(ctx context.Context) (int, error) {
	return d.Advance(ctx, d.doc+1)
}

func (d *docIdSetIteratorAll) Advance(ctx context.Context, target int) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	d.doc = target
	if d.doc >= d.maxDoc {
		d.doc = NO_MORE_DOCS
		return 0, io.EOF
	}
	return d.doc, nil
}

func (d *docIdSetIteratorAll) Cost() int64 {
	return int64(d.maxDoc)
}

var _ DocIdSetIterator = &emptyDocIdSetIterator{}

func GetEmptyDocIdSetIterator() DocIdSetIterator {
	return &emptyDocIdSetIterator{}
}

type emptyDocIdSetIterator struct {
	exhausted bool
}

func (e *emptyDocIdSetIterator) DocID() int {
	if e.exhausted {
		return NO_MORE_DOCS
	}
	return -1
}

func (e *emptyDocIdSetIterator) NextDoc(ctx context.Context) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	e.exhausted = true
	return 0, io.EOF
}

func (e *emptyDocIdSetIterator) Advance(ctx context.Context, target int) (int, error) {
	e.exhausted = true
	return 0, io.EOF
}

func (e *emptyDocIdSetIterator) SlowAdvance(ctx context.Context, target int) (int, error) {
	return SlowAdvanceWithContext(ctx, e, target)
}

func (e *emptyDocIdSetIterator) Cost() int64 {
	return 0
}
