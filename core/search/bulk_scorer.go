package search

import (
	"github.com/geange/lucene-go/core/util"
	"io"
	"math"
)

// BulkScorer This class is used to Score a range of documents at once, and is returned by Weight.bulkScorer.
// Only queries that have a more optimized means of scoring across a range of documents need to override this.
// Otherwise, a default implementation is wrapped around the Scorer returned by Weight.scorer.
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

type BulkScorerDefault struct {
	FnScoreRange func(collector LeafCollector, acceptDocs util.Bits, min, max int) (int, error)
	FnCost       func() int64
}

func (b *BulkScorerDefault) Score(collector LeafCollector, acceptDocs util.Bits) error {
	_, err := b.FnScoreRange(collector, acceptDocs, 0, math.MaxInt)
	if err != io.EOF {
		panic("")
	}
	return nil
}

func (b *BulkScorerDefault) ScoreRange(collector LeafCollector, acceptDocs util.Bits, min, max int) (int, error) {
	return b.FnScoreRange(collector, acceptDocs, min, max)
}

func (b *BulkScorerDefault) Cost() int64 {
	return b.FnCost()
}
