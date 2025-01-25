package search

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

type baseLeafCollector struct {
}

func (*baseLeafCollector) CompetitiveIterator() (types.DocIdSetIterator, error) {
	return nil, nil
}

type FilterLeafCollector struct {
	in index.LeafCollector
}

var _ index.LeafCollector = &LeafCollectorAnon{}

type LeafCollectorAnon struct {
	FnSetScorer           func(scorer index.Scorable) error
	FnCollect             func(ctx context.Context, doc int) error
	FnCompetitiveIterator func() (types.DocIdSetIterator, error)
}

func (l *LeafCollectorAnon) SetScorer(scorer index.Scorable) error {
	return l.FnSetScorer(scorer)
}

func (l *LeafCollectorAnon) Collect(ctx context.Context, doc int) error {
	return l.FnCollect(ctx, doc)
}

func (l *LeafCollectorAnon) CompetitiveIterator() (types.DocIdSetIterator, error) {
	return l.FnCompetitiveIterator()
}
