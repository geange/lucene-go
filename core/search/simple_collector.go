package search

import "github.com/geange/lucene-go/core/index"

// SimpleCollector Base Collector implementation that is used to collect all contexts.
type SimpleCollector interface {
	Collector
	LeafCollector

	// DoSetNextReader This method is called before collecting context.
	DoSetNextReader(context *index.LeafReaderContext) error
}

type SimpleCollectorExtra interface {
	DoSetNextReader(context *index.LeafReaderContext) error
	SetScorer(scorer Scorable) error
	Collect(doc int) error
}

type SimpleCollectorImp struct {
	*LeafCollectorImp

	SimpleCollectorExtra
}

func NewSimpleCollectorImp(extra SimpleCollectorExtra) *SimpleCollectorImp {
	return &SimpleCollectorImp{SimpleCollectorExtra: extra}
}

func (s *SimpleCollectorImp) GetLeafCollector(context *index.LeafReaderContext) (LeafCollector, error) {
	err := s.DoSetNextReader(context)
	if err != nil {
		return nil, err
	}
	return s, nil
}
