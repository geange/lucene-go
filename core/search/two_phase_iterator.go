package search

import (
	"github.com/geange/lucene-go/core/index"
)

// TwoPhaseIterator Returned by Scorer.twoPhaseIterator() to expose an approximation of a DocIdSetIterator. When the approximation()'s DocIdSetIterator.nextDoc() or DocIdSetIterator.advance(int) return, matches() needs to be checked in order to know whether the returned doc ID actually matches.
type TwoPhaseIterator interface {
	Approximation() index.DocIdSetIterator

	// Matches Return whether the current doc ID that approximation() is on matches. This method should only be called when the iterator is positioned -- ie. not when DocIdSetIterator.docID() is -1 or DocIdSetIterator.NO_MORE_DOCS -- and at most once.
	Matches() (bool, error)

	// MatchCost An estimate of the expected cost to determine that a single document matches(). This can be called before iterating the documents of approximation(). Returns an expected cost in number of simple operations like addition, multiplication, comparing two numbers and indexing an array. The returned value must be positive.
	MatchCost() float64
}
