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
}

type SimpleCollectorImp struct {
	SimpleCollectorExtra
}

func (s *SimpleCollectorImp) GetLeafCollector(context *index.LeafReaderContext) (LeafCollector, error) {
	err := s.DoSetNextReader(context)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SimpleCollectorImp) SetScorer(scorer Scorable) error {
	return nil
}

func (s *SimpleCollectorImp) Collect(doc int) error {
	return nil
}
