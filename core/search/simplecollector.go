package search

import (
	"context"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/interface/search"
)

// SimpleCollector
// Base Collector implementation that is used to collect all contexts.
type SimpleCollector interface {
	search.Collector
	search.LeafCollector

	// DoSetNextReader
	// This method is called before collecting context.
	DoSetNextReader(context index.LeafReaderContext) error
}

type SimpleCollectorSPI interface {
	DoSetNextReader(context index.LeafReaderContext) error
	SetScorer(scorer search.Scorable) error
	Collect(ctx context.Context, doc int) error
}

type BaseSimpleCollector struct {
	*baseLeafCollector

	SimpleCollectorSPI
}

func NewSimpleCollector(spi SimpleCollectorSPI) *BaseSimpleCollector {
	return &BaseSimpleCollector{SimpleCollectorSPI: spi}
}

func (s *BaseSimpleCollector) GetLeafCollector(ctx context.Context, readerContext index.LeafReaderContext) (search.LeafCollector, error) {
	if err := s.DoSetNextReader(readerContext); err != nil {
		return nil, err
	}
	return s, nil
}
