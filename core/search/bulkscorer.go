package search

import (
	"math"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util"
)

type ScoreRange func(collector index.LeafCollector, acceptDocs util.Bits, from, to int) (int, error)
type ScoreCost func() int64

type BulkScorer interface {
	GetScorer() (ScoreRange, ScoreCost)
}

type BaseBulkScorer struct {
	FnScoreRange func(collector index.LeafCollector, acceptDocs util.Bits, min, max int) (int, error)
	FnCost       func() int64
}

func (b *BaseBulkScorer) Score(collector index.LeafCollector, acceptDocs util.Bits, minDoc, maxDoc int) (int, error) {
	if minDoc < 0 && maxDoc < 0 {
		minDoc = 0
		maxDoc = math.MaxInt32
	}
	return b.FnScoreRange(collector, acceptDocs, minDoc, maxDoc)
}

func (b *BaseBulkScorer) Cost() int64 {
	return b.FnCost()
}
