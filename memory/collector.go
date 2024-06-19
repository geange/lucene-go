package memory

import (
	"context"
	"github.com/geange/lucene-go/core/interface/index"
	search2 "github.com/geange/lucene-go/core/interface/search"
	"github.com/geange/lucene-go/core/search"
)

var _ search.SimpleCollector = &simpleCollector{}

type simpleCollector struct {
	*search.BaseSimpleCollector

	scorer search2.Scorable
	scores []float64
}

func newSimpleCollector(scores []float64) *simpleCollector {
	collector := &simpleCollector{
		BaseSimpleCollector: nil,
		scorer:              nil,
		scores:              scores,
	}
	collector.BaseSimpleCollector = search.NewSimpleCollector(collector)
	return collector
}

func (s *simpleCollector) ScoreMode() search2.ScoreMode {
	return search.COMPLETE
}

func (s *simpleCollector) Collect(ctx context.Context, doc int) error {
	var err error
	score, err := s.scorer.Score()
	if err != nil {
		return err
	}
	s.scores[0] = score
	return err
}

func (s *simpleCollector) DoSetNextReader(_ index.LeafReaderContext) error {
	return nil
}

func (s *simpleCollector) SetScorer(scorer search2.Scorable) error {
	s.scorer = scorer
	return nil
}
