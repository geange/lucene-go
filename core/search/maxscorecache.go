package search

import (
	"github.com/geange/lucene-go/core/index"
	index2 "github.com/geange/lucene-go/core/interface/index"
)

// MaxScoreCache Compute maximum scores based on Impacts and keep them in a cache in order not to run
// expensive similarity score computations multiple times on the same data.
type MaxScoreCache struct {
	impactsSource     index2.ImpactsSource
	scorer            index.SimScorer
	maxScoreCache     []float64
	maxScoreCacheUpTo []int
}

func (c *MaxScoreCache) GetMaxScoreForLevel(level int) (float64, error) {
	impacts, err := c.impactsSource.GetImpacts()
	if err != nil {
		return 0, err
	}
	c.ensureCacheSize(level + 1)
	levelUpTo := impacts.GetDocIdUpTo(level)
	if c.maxScoreCacheUpTo[level] < levelUpTo {
		c.maxScoreCache[level] = c.computeMaxScore(impacts.GetImpacts(level))
		c.maxScoreCacheUpTo[level] = levelUpTo
	}
	return c.maxScoreCache[level], nil
}

// GetSkipUpTo
// Return the an inclusive upper bound of documents that all have a score
// that is less than minScore, or -1 if the current document may be competitive.
func (c *MaxScoreCache) GetSkipUpTo(minScore float64) (int, error) {
	impacts, err := c.impactsSource.GetImpacts()
	if err != nil {
		return 0, err
	}
	level, err := c.getSkipLevel(impacts, minScore)
	if err != nil {
		return 0, err
	}
	if level == -1 {
		return -1, nil
	}
	return impacts.GetDocIdUpTo(level), nil
}

func (c *MaxScoreCache) ensureCacheSize(size int) {
	if len(c.maxScoreCache) < size {
		oldLength := len(c.maxScoreCache)
		if oldLength < size {
			c.maxScoreCache = append(c.maxScoreCache, make([]float64, size-oldLength)...)
		}
		tmp := make([]int, len(c.maxScoreCache))

		copy(tmp, c.maxScoreCacheUpTo)
		c.maxScoreCacheUpTo = tmp

		for i := oldLength; i < len(c.maxScoreCacheUpTo); i++ {
			c.maxScoreCacheUpTo[i] = -1
		}
	}
}

func (c *MaxScoreCache) computeMaxScore(impacts []index2.Impact) float64 {
	maxScore := float64(0)
	for _, impact := range impacts {
		maxScore = max(c.scorer.Score(float64(impact.GetFreq()), impact.GetNorm()), maxScore)
	}
	return maxScore
}

// Return the maximum level at which scores are all less than minScore, or -1 if none.
func (c *MaxScoreCache) getSkipLevel(impacts index2.Impacts, minScore float64) (int, error) {
	numLevels := impacts.NumLevels()
	for level := 0; level < numLevels; level++ {
		forLevel, err := c.GetMaxScoreForLevel(level)
		if err != nil {
			return 0, err
		}
		if forLevel >= minScore {
			return level - 1, nil
		}
	}
	return numLevels - 1, nil
}

// GetLevel
// Return the first level that includes all doc IDs up to upTo, or -1 if there is no such level.
func (c *MaxScoreCache) GetLevel(upTo int) (int, error) {
	impacts, err := c.impactsSource.GetImpacts()
	if err != nil {
		return 0, err
	}
	level := 0
	for numLevels := impacts.NumLevels(); level < numLevels; level++ {
		impactsUpTo := impacts.GetDocIdUpTo(level)
		if upTo <= impactsUpTo {
			return level, nil
		}
	}
	return -1, nil
}

func NewMaxScoreCache(impactsSource index2.ImpactsSource, scorer index.SimScorer) *MaxScoreCache {
	return &MaxScoreCache{
		impactsSource:     impactsSource,
		scorer:            scorer,
		maxScoreCache:     make([]float64, 0),
		maxScoreCacheUpTo: make([]int, 0),
	}
}
