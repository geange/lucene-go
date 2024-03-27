package search

import (
	"errors"
	"github.com/geange/lucene-go/core/util"
	"io"
	"math"
)

// BulkScorer
// This class is used to Score a range of documents at once, and is returned by Weight.bulkScorer.
// Only queries that have a more optimized means of scoring across a range of documents need to override this.
// Otherwise, a default implementation is wrapped around the Scorer returned by Weight.scorer.
//
// GPT3.5：
// 这个类用于一次对一系列文档进行评分，它是由Weight.bulkScorer返回的。
// 只有那些在一系列文档上有更优化的评分方法的查询才需要覆盖它。
// 否则，会使用默认实现，该实现会封装在Weight.scorer返回的Scorer周围。
type BulkScorer interface {
	// Score Scores and collects all matching documents.
	// Params: 	collector – The collector to which all matching documents are passed.
	//			acceptDocs – Bits that represents the allowed documents to match, or null if they are all allowed to match.
	Score(collector LeafCollector, acceptDocs util.Bits) error

	// ScoreRange
	// Params:
	// 		collector – The collector to which all matching documents are passed.
	// 		acceptDocs – Bits that represents the allowed documents to match, or null if they are all allowed to match.
	// 		min – Score starting at, including, this document
	// 		max – Score up to, but not including, this doc
	// Returns: an under-estimation of the next matching doc after max
	ScoreRange(collector LeafCollector, acceptDocs util.Bits, min, max int) (int, error)

	// Cost Same as DocIdSetIterator.cost() for bulk scorers.
	Cost() int64
}

type BulkScorerSPI interface {
	ScoreRange(collector LeafCollector, acceptDocs util.Bits, min, max int) (int, error)
	Cost() int64
}

type BaseBulkScorer struct {
	FnScoreRange func(collector LeafCollector, acceptDocs util.Bits, min, max int) (int, error)
	FnCost       func() int64
}

func (b *BaseBulkScorer) Score(collector LeafCollector, acceptDocs util.Bits) error {
	if _, err := b.FnScoreRange(collector, acceptDocs, 0, math.MaxInt); errors.Is(err, io.EOF) {
		return nil
	} else {
		return err
	}
}

func (b *BaseBulkScorer) ScoreRange(collector LeafCollector, acceptDocs util.Bits, min, max int) (int, error) {
	return b.FnScoreRange(collector, acceptDocs, min, max)
}

func (b *BaseBulkScorer) Cost() int64 {
	return b.FnCost()
}

var _ BulkScorer = &BulkScorerAnon{}

type BulkScorerAnon struct {
	FnScore      func(collector LeafCollector, acceptDocs util.Bits) error
	FnScoreRange func(collector LeafCollector, acceptDocs util.Bits, min, max int) (int, error)
	FnCost       func() int64
}

func (b *BulkScorerAnon) Score(collector LeafCollector, acceptDocs util.Bits) error {
	if b.FnScore != nil {
		return b.FnScore(collector, acceptDocs)
	}
	if _, err := b.ScoreRange(collector, acceptDocs, 0, math.MaxInt32); err != nil {
		return err
	}
	return nil
}

func (b *BulkScorerAnon) ScoreRange(collector LeafCollector, acceptDocs util.Bits, min, max int) (int, error) {
	return b.FnScoreRange(collector, acceptDocs, min, max)
}

func (b *BulkScorerAnon) Cost() int64 {
	return b.FnCost()
}
