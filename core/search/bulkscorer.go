package search

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util"
	"io"
	"math"
)

type BulkScorerSPI interface {
	ScoreRange(collector index.LeafCollector, acceptDocs util.Bits, min, max int) (int, error)
	Cost() int64
}

type BaseBulkScorer struct {
	FnScoreRange func(ctx context.Context, collector index.LeafCollector, acceptDocs util.Bits, min, max int) (int, error)
	FnCost       func() int64
}

func (b *BaseBulkScorer) Score(ctx context.Context, collector index.LeafCollector, acceptDocs util.Bits) error {
	if _, err := b.FnScoreRange(ctx, collector, acceptDocs, 0, math.MaxInt); errors.Is(err, io.EOF) {
		return nil
	} else {
		return err
	}
}

func (b *BaseBulkScorer) ScoreRange(ctx context.Context, collector index.LeafCollector, acceptDocs util.Bits, min, max int) (int, error) {
	return b.FnScoreRange(ctx, collector, acceptDocs, min, max)
}

func (b *BaseBulkScorer) Cost() int64 {
	return b.FnCost()
}

var _ index.BulkScorer = &BulkScorerAnon{}

type BulkScorerAnon struct {
	FnScore      func(collector index.LeafCollector, acceptDocs util.Bits) error
	FnScoreRange func(collector index.LeafCollector, acceptDocs util.Bits, min, max int) (int, error)
	FnCost       func() int64
}

func (b *BulkScorerAnon) Score(ctx context.Context, collector index.LeafCollector, acceptDocs util.Bits) error {
	if b.FnScore != nil {
		return b.FnScore(collector, acceptDocs)
	}
	if _, err := b.ScoreRange(ctx, collector, acceptDocs, 0, math.MaxInt32); err != nil {
		return err
	}
	return nil
}

func (b *BulkScorerAnon) ScoreRange(ctx context.Context, collector index.LeafCollector, acceptDocs util.Bits, min, max int) (int, error) {
	return b.FnScoreRange(collector, acceptDocs, min, max)
}

func (b *BulkScorerAnon) Cost() int64 {
	return b.FnCost()
}
