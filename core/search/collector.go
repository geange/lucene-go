package search

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
)

// SimpleCollector
// Base Collector implementation that is used to collect all contexts.
type SimpleCollector interface {
	index.Collector
	index.LeafCollector

	// DoSetNextReader
	// This method is called before collecting context.
	DoSetNextReader(context index.LeafReaderContext) error
}

type SimpleCollectorSPI interface {
	DoSetNextReader(context index.LeafReaderContext) error
	SetScorer(scorer index.Scorable) error
	Collect(ctx context.Context, doc int) error
}

type BaseSimpleCollector struct {
	*baseLeafCollector

	SimpleCollectorSPI
}

func NewSimpleCollector(spi SimpleCollectorSPI) *BaseSimpleCollector {
	return &BaseSimpleCollector{SimpleCollectorSPI: spi}
}

func (s *BaseSimpleCollector) GetLeafCollector(ctx context.Context, readerContext index.LeafReaderContext) (index.LeafCollector, error) {
	if err := s.DoSetNextReader(readerContext); err != nil {
		return nil, err
	}
	return s, nil
}

var _ SimpleCollector = &TotalHitCountCollector{}

type TotalHitCountCollector struct {
	*BaseSimpleCollector

	totalHits int
}

func NewTotalHitCountCollector() *TotalHitCountCollector {
	collector := &TotalHitCountCollector{}
	collector.BaseSimpleCollector = NewSimpleCollector(collector)
	return collector
}

func (t *TotalHitCountCollector) Collect(ctx context.Context, doc int) error {
	t.totalHits++
	return nil
}

func (t *TotalHitCountCollector) ScoreMode() index.ScoreMode {
	return COMPLETE_NO_SCORES
}

func (t *TotalHitCountCollector) GetTotalHits() int {
	return t.totalHits
}
