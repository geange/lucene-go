package search

import (
	"github.com/geange/lucene-go/core/index"
)

// Scorer Expert: Common scoring functionality for different types of queries.
// 不同类型查询的通用评分功能。
//
// A Scorer exposes an iterator() over documents matching a query in increasing order of doc Id.
// 计分器暴露一个迭代器，这个迭代器按照文档id递增顺序
//
// Document scores are computed using a given Similarity implementation.
// NOTE: The values Float.Nan, Float.NEGATIVE_INFINITY and Float.POSITIVE_INFINITY are not valid scores.
// Certain collectors (eg TopScoreDocCollector) will not properly collect hits with these scores.
type Scorer interface {
	Scorable

	// GetWeight returns parent Weight
	GetWeight() Weight

	// Iterator Return a DocIdSetIterator over matching documents. The returned iterator will either
	// be positioned on -1 if no documents have been scored yet, DocIdSetIterator.NO_MORE_DOCS if all
	// documents have been scored already, or the last document id that has been scored otherwise.
	// The returned iterator is a view: calling this method several times will return iterators
	// that have the same state.
	Iterator() index.DocIdSetIterator

	// TwoPhaseIterator Optional method: Return a TwoPhaseIterator view of this Scorer. A return value
	// of null indicates that two-phase iteration is not supported. Note that the returned
	// TwoPhaseIterator's approximation must advance synchronously with the iterator(): advancing
	// the approximation must advance the iterator and vice-versa. Implementing this method is
	// typically useful on Scorers that have a high per-document overhead in order to confirm
	// matches. The default implementation returns null.
	TwoPhaseIterator() TwoPhaseIterator

	// GetMaxScore Return the maximum score that documents between the last target that this iterator
	// was shallow-advanced to included and upTo included.
	GetMaxScore(upTo int) (float64, error)
}
