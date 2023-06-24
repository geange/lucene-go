package search

import (
	"github.com/geange/lucene-go/core/index"
	"io"
)

// TwoPhaseIterator
// Returned by Scorer.TwoPhaseIterator() to expose an approximation of a DocIdSetIterator.
// When the approximation()'s DocIdSetIterator.nextDoc() or DocIdSetIterator.advance(int) return,
// matches() needs to be checked in order to know whether the returned doc ID actually matches.
//
// # GPT3.5
//
// 在Lucene中，`TwoPhaseIterator`是一个用于执行两阶段迭代的工具类。
// 它可以与Scorer一起使用，用于更高效地过滤和评分匹配文档。
//
// 在搜索过程中，通常会使用一个Scorer来进行文档匹配，并对匹配的文档进行评分。然而，有时候在进行文档匹配之前，
// 可以使用一些更快速的方法来过滤掉不符合条件的文档，从而减少评分操作的开销。
//
// `TwoPhaseIterator`类就提供了这样的功能。它通过两个阶段的迭代来实现过滤和评分的分离。
//
// 在第一阶段，`TwoPhaseIterator`会对文档进行快速的过滤操作，根据一些预先计算的条件（例如，布尔表达式或位集合），
// 判断文档是否可能匹配查询条件。这个过滤操作通常比完全匹配文档的评分操作更快。
//
// 在第二阶段，对于通过第一阶段过滤的文档，`TwoPhaseIterator`会将这些文档传递给实际的Scorer进行详细的匹配和评分操作。
//
// 使用`TwoPhaseIterator`的好处是，它可以减少不必要的评分操作，只对通过过滤的文档进行实际的匹配和评分，从而提高搜索性能。
//
// `TwoPhaseIterator`类主要包含以下方法：
//
// 1. `approximation()`：返回用于快速过滤的近似评分器（approximation scorer）。
//
// 2. `matches()`：在第一阶段中，检查当前文档是否匹配查询条件。
//
// 3. `matchCost()`：返回第一阶段中过滤操作的成本。用于估算在第一阶段过滤后剩余的文档数量。
//
// 通过使用`TwoPhaseIterator`，可以在搜索过程中根据具体需求进行过滤和评分的优化，提高搜索性能并降低开销。
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
