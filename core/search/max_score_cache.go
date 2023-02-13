package search

import (
	"github.com/geange/lucene-go/core/index"
)

// MaxScoreCache Compute maximum scores based on Impacts and keep them in a cache in order not to run
// expensive similarity score computations multiple times on the same data.
type MaxScoreCache struct {
	impactsSource     index.ImpactsSource
	scorer            index.SimScorer
	maxScoreCache     []float64
	maxScoreCacheUpTo []int
}

func NewMaxScoreCache(impactsSource index.ImpactsSource, scorer index.SimScorer) *MaxScoreCache {
	return &MaxScoreCache{
		impactsSource:     impactsSource,
		scorer:            scorer,
		maxScoreCache:     make([]float64, 0),
		maxScoreCacheUpTo: make([]int, 0),
	}
}
