package search

import (
	"github.com/geange/lucene-go/core/index"
	"io"
)

// TwoPhaseIterator Returned by Scorer.twoPhaseIterator() to expose an approximation of a DocIdSetIterator. When the approximation()'s DocIdSetIterator.nextDoc() or DocIdSetIterator.advance(int) return, matches() needs to be checked in order to know whether the returned doc ID actually matches.
type TwoPhaseIterator interface {
	Approximation() index.DocIdSetIterator

	// Matches
	// Return whether the current doc ID that approximation() is on matches.
	// This method should only be called when the iterator is positioned -- ie. not when DocIdSetIterator.docID() is -1 or DocIdSetIterator.NO_MORE_DOCS -- and at most once.
	Matches() (bool, error)

	// MatchCost
	// An estimate of the expected cost to determine that a single document matches().
	// This can be called before iterating the documents of approximation().
	// Returns an expected cost in number of simple operations like addition, multiplication, comparing two numbers and indexing an array. The returned value must be positive.
	MatchCost() float64
}

func AsDocIdSetIterator(twoPhaseIterator TwoPhaseIterator) index.DocIdSetIterator {
	return NewTwoPhaseIteratorAsDocIdSetIterator(twoPhaseIterator)
}

var _ index.DocIdSetIterator = &TwoPhaseIteratorAsDocIdSetIterator{}

type TwoPhaseIteratorAsDocIdSetIterator struct {
	twoPhaseIterator TwoPhaseIterator
	approximation    index.DocIdSetIterator
}

func NewTwoPhaseIteratorAsDocIdSetIterator(twoPhaseIterator TwoPhaseIterator) *TwoPhaseIteratorAsDocIdSetIterator {
	return &TwoPhaseIteratorAsDocIdSetIterator{
		twoPhaseIterator: twoPhaseIterator,
		approximation:    twoPhaseIterator.Approximation(),
	}
}

func (t *TwoPhaseIteratorAsDocIdSetIterator) DocID() int {
	return t.approximation.DocID()
}

func (t *TwoPhaseIteratorAsDocIdSetIterator) NextDoc() (int, error) {
	doc, err := t.approximation.NextDoc()
	if err != nil {
		return 0, err
	}
	return t.doNext(doc)
}

func (t *TwoPhaseIteratorAsDocIdSetIterator) Advance(target int) (int, error) {
	doc, err := t.approximation.Advance(target)
	if err != nil {
		return 0, err
	}
	return t.doNext(doc)
}

func (t *TwoPhaseIteratorAsDocIdSetIterator) SlowAdvance(target int) (int, error) {
	return index.SlowAdvance(t, target)
}

func (t *TwoPhaseIteratorAsDocIdSetIterator) Cost() int64 {
	return t.approximation.Cost()
}

func (t *TwoPhaseIteratorAsDocIdSetIterator) doNext(doc int) (int, error) {
	for {
		if doc == index.NO_MORE_DOCS {
			return 0, io.EOF
		}

		matches, err := t.twoPhaseIterator.Matches()
		if err != nil {
			return 0, err
		}
		if matches {
			return doc, nil
		}

		doc = t.approximation.DocID()
	}
}

func UnwrapIterator(iterator index.DocIdSetIterator) TwoPhaseIterator {
	if v, ok := iterator.(*TwoPhaseIteratorAsDocIdSetIterator); ok {
		return v.twoPhaseIterator
	}
	return nil
}
