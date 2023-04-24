package memory

import (
	"context"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search"
)

var _ search.SimpleCollector = &simpleCollector{}

type simpleCollector struct {
	*search.SimpleCollectorImp
	scorer search.Scorable
	scores []float64
}

func newSimpleCollector(scores []float64) *simpleCollector {
	collector := &simpleCollector{
		SimpleCollectorImp: nil,
		scorer:             nil,
		scores:             scores,
	}
	collector.SimpleCollectorImp = search.NewSimpleCollectorImp(collector)
	return collector
}

func (s *simpleCollector) ScoreMode() *search.ScoreMode {
	return search.COMPLETE
}

func (s *simpleCollector) Collect(ctx context.Context, doc int) error {
	var err error
	score, err := s.scorer.Score()
	if err != nil {
		return err
	}
	s.scores[0] = float64(score)
	return err
}

func (s *simpleCollector) DoSetNextReader(context *index.LeafReaderContext) error {
	return nil
}

func (s *simpleCollector) SetScorer(scorer search.Scorable) error {
	s.scorer = scorer
	return nil
}
