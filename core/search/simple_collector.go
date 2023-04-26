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

type SimpleCollectorDefault struct {
	*LeafCollectorDefault

	SimpleCollectorSPI
}

func NewSimpleCollectorDefault(spi SimpleCollectorSPI) *SimpleCollectorDefault {
	return &SimpleCollectorDefault{SimpleCollectorSPI: spi}
}

func (s *SimpleCollectorDefault) GetLeafCollector(ctx context.Context, leafCtx *index.LeafReaderContext) (LeafCollector, error) {
	err := s.DoSetNextReader(leafCtx)
	if err != nil {
		return nil, err
	}
	return s, nil
}
