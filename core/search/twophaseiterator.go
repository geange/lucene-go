package search

import (
	"github.com/geange/lucene-go/core/interface/search"
	"github.com/geange/lucene-go/core/types"
	"io"
)

func AsDocIdSetIterator(twoPhaseIterator search.TwoPhaseIterator) types.DocIdSetIterator {
	return &twoPhaseIteratorAsDocIdSetIterator{
		twoPhaseIterator: twoPhaseIterator,
		approximation:    twoPhaseIterator.Approximation(),
	}
}

var _ types.DocIdSetIterator = &twoPhaseIteratorAsDocIdSetIterator{}

type twoPhaseIteratorAsDocIdSetIterator struct {
	twoPhaseIterator search.TwoPhaseIterator
	approximation    types.DocIdSetIterator
}

func (t *twoPhaseIteratorAsDocIdSetIterator) DocID() int {
	return t.approximation.DocID()
}

func (t *twoPhaseIteratorAsDocIdSetIterator) NextDoc() (int, error) {
	doc, err := t.approximation.NextDoc()
	if err != nil {
		return 0, err
	}
	return t.doNext(doc)
}

func (t *twoPhaseIteratorAsDocIdSetIterator) Advance(target int) (int, error) {
	doc, err := t.approximation.Advance(target)
	if err != nil {
		return 0, err
	}
	return t.doNext(doc)
}

func (t *twoPhaseIteratorAsDocIdSetIterator) SlowAdvance(target int) (int, error) {
	return types.SlowAdvance(t, target)
}

func (t *twoPhaseIteratorAsDocIdSetIterator) Cost() int64 {
	return t.approximation.Cost()
}

func (t *twoPhaseIteratorAsDocIdSetIterator) doNext(doc int) (int, error) {
	for {
		if doc == types.NO_MORE_DOCS {
			return 0, io.EOF
		}

		isMatch, err := t.twoPhaseIterator.Matches()
		if err != nil {
			return 0, err
		}
		if isMatch {
			return doc, nil
		}

		doc = t.approximation.DocID()
	}
}

func UnwrapIterator(iterator types.DocIdSetIterator) search.TwoPhaseIterator {
	if v, ok := iterator.(*twoPhaseIteratorAsDocIdSetIterator); ok {
		return v.twoPhaseIterator
	}
	return nil
}
