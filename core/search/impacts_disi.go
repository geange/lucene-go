package search

import (
	"github.com/geange/lucene-go/core/index"
	"math"
)

var _ index.DocIdSetIterator = &ImpactsDISI{}

// ImpactsDISI DocIdSetIterator that skips non-competitive docs thanks to the indexed impacts.
// Call setMinCompetitiveScore(float) in order to give this iterator the ability to skip
// low-scoring documents.
type ImpactsDISI struct {
	in                  index.DocIdSetIterator
	impactsSource       index.ImpactsSource
	maxScoreCache       *MaxScoreCache
	globalMaxScore      float64
	minCompetitiveScore float64
	upTo                int
	maxScore            float64
}

func NewImpactsDISI(in index.DocIdSetIterator, impactsSource index.ImpactsSource, scorer SimScorer) *ImpactsDISI {
	return &ImpactsDISI{
		in:             in,
		impactsSource:  impactsSource,
		maxScoreCache:  NewMaxScoreCache(impactsSource, scorer),
		globalMaxScore: scorer.Score(math.MaxFloat32, 1),
	}
}

func (d *ImpactsDISI) setMinCompetitiveScore(score float64) error {
	panic("")
}

func (d *ImpactsDISI) getMaxScore(to int) (float64, error) {
	panic("")
}

func (d *ImpactsDISI) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (d *ImpactsDISI) NextDoc() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (d *ImpactsDISI) Advance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (d *ImpactsDISI) SlowAdvance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (d *ImpactsDISI) Cost() int64 {
	//TODO implement me
	panic("implement me")
}
