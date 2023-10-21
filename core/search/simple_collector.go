package search

import (
	"context"
	"github.com/geange/lucene-go/core/index"
)

// SimpleCollector
// Base Collector implementation that is used to collect all contexts.
type SimpleCollector interface {
	Collector
	LeafCollector

	// DoSetNextReader This method is called before collecting context.
	DoSetNextReader(context *index.LeafReaderContext) error
}

type SimpleCollectorSPI interface {
	DoSetNextReader(context *index.LeafReaderContext) error
	SetScorer(scorer Scorable) error
	Collect(ctx context.Context, doc int) error
}

type DefSimpleCollector struct {
	*DefLeafCollector

	SimpleCollectorSPI
}

func NewSimpleCollector(spi SimpleCollectorSPI) *DefSimpleCollector {
	return &DefSimpleCollector{SimpleCollectorSPI: spi}
}

func (s *DefSimpleCollector) GetLeafCollector(_ context.Context, ctx *index.LeafReaderContext) (LeafCollector, error) {
	err := s.DoSetNextReader(ctx)
	if err != nil {
		return nil, err
	}
	return s, nil
}
