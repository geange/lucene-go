package search

import (
	"errors"
	"github.com/geange/lucene-go/core/interface/search"
	"github.com/geange/lucene-go/core/util"
	"io"
	"math"
)

type BulkScorerSPI interface {
	ScoreRange(collector search.LeafCollector, acceptDocs util.Bits, min, max int) (int, error)
	Cost() int64
}

type BaseBulkScorer struct {
	FnScoreRange func(collector search.LeafCollector, acceptDocs util.Bits, min, max int) (int, error)
	FnCost       func() int64
}

func (b *BaseBulkScorer) Score(collector search.LeafCollector, acceptDocs util.Bits) error {
	if _, err := b.FnScoreRange(collector, acceptDocs, 0, math.MaxInt); errors.Is(err, io.EOF) {
		return nil
	} else {
		return err
	}
}

func (b *BaseBulkScorer) ScoreRange(collector search.LeafCollector, acceptDocs util.Bits, min, max int) (int, error) {
	return b.FnScoreRange(collector, acceptDocs, min, max)
}

func (b *BaseBulkScorer) Cost() int64 {
	return b.FnCost()
}

var _ search.BulkScorer = &BulkScorerAnon{}

type BulkScorerAnon struct {
	FnScore      func(collector search.LeafCollector, acceptDocs util.Bits) error
	FnScoreRange func(collector search.LeafCollector, acceptDocs util.Bits, min, max int) (int, error)
	FnCost       func() int64
}

func (b *BulkScorerAnon) Score(collector search.LeafCollector, acceptDocs util.Bits) error {
	if b.FnScore != nil {
		return b.FnScore(collector, acceptDocs)
	}
	if _, err := b.ScoreRange(collector, acceptDocs, 0, math.MaxInt32); err != nil {
		return err
	}
	return nil
}

func (b *BulkScorerAnon) ScoreRange(collector search.LeafCollector, acceptDocs util.Bits, min, max int) (int, error) {
	return b.FnScoreRange(collector, acceptDocs, min, max)
}

func (b *BulkScorerAnon) Cost() int64 {
	return b.FnCost()
}
