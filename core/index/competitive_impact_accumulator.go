package index

import (
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/emirpasic/gods/utils"
	"github.com/geange/lucene-go/core/util"
	"math"
)

// CompetitiveImpactAccumulator This class accumulates the (freq, norm) pairs that may produce competitive scores.
type CompetitiveImpactAccumulator struct {
	// We speed up accumulation for common norm values with this array that maps
	// norm values in -128..127 to the maximum frequency observed for these norm
	// values
	maxFreqs []int

	// This TreeSet stores competitive (freq,norm) pairs for norm values that fall
	// outside of -128..127. It is always empty with the default similarity, which
	// encodes norms as bytes.
	otherFreqNormPairs *treeset.Set
}

func NewCompetitiveImpactAccumulator() *CompetitiveImpactAccumulator {
	return &CompetitiveImpactAccumulator{
		maxFreqs:           make([]int, 256),
		otherFreqNormPairs: treeset.NewWith(ImpactComparator),
	}
}

func (c *CompetitiveImpactAccumulator) Clear() {
	for i := range c.maxFreqs {
		c.maxFreqs[i] = 0
	}
	c.otherFreqNormPairs.Clear()
}

// Add Accumulate a (freq,norm) pair, updating this structure if there is no equivalent or more competitive entry already.
func (c *CompetitiveImpactAccumulator) Add(freq int, norm int64) {
	if norm >= math.MinInt8 && norm <= math.MaxInt8 {
		index := uint(norm)
		c.maxFreqs[index] = util.Max(c.maxFreqs[index], freq)
		return
	}
	c.add(NewImpact(freq, norm), c.otherFreqNormPairs)
}

func (c *CompetitiveImpactAccumulator) add(newEntry *Impact, freqNormPairs *treeset.Set) {
	_, next := freqNormPairs.Find(func(index int, value interface{}) bool {
		if ImpactComparator(*newEntry, value) <= 0 {
			return true
		}
		return false
	})
	if next == nil {
		freqNormPairs.Add(*newEntry)
	} else if utils.Int64Comparator(next.(Impact).Norm, newEntry.Norm) <= 0 {
		return
	} else {
		freqNormPairs.Add(*newEntry)
	}

	iterator := freqNormPairs.Iterator()

	iterator.NextTo(func(index int, value interface{}) bool {
		if ImpactComparator(*newEntry, value) > 0 {
			if utils.Int64Comparator(newEntry.Norm, value.(Impact).Norm) >= 0 {
				freqNormPairs.Remove(value)
			}
			return true
		}

		return false
	})
}

// AddAll Merge acc into this.
func (c *CompetitiveImpactAccumulator) AddAll(acc *CompetitiveImpactAccumulator) {
	maxFreqs := c.maxFreqs
	otherMaxFreqs := acc.maxFreqs

	for i := 0; i < len(maxFreqs); i++ {
		maxFreqs[i] = util.Max(maxFreqs[i], otherMaxFreqs[i])
	}

	for _, v := range acc.otherFreqNormPairs.Values() {
		entry := v.(Impact)
		c.add(&entry, c.otherFreqNormPairs)
	}
}

// GetCompetitiveFreqNormPairs Get the set of competitive freq and norm pairs, ordered by increasing freq and norm.
func (c *CompetitiveImpactAccumulator) GetCompetitiveFreqNormPairs() []*Impact {
	impacts := make([]*Impact, 0)
	maxFreqForLowerNorms := 0
	for i := range c.maxFreqs {
		maxFreq := c.maxFreqs[i]
		if maxFreq > maxFreqForLowerNorms {
			impacts = append(impacts, NewImpact(maxFreq, int64(i)))
			maxFreqForLowerNorms = maxFreq
		}
	}

	if c.otherFreqNormPairs.Size() == 0 {
		// Common case: all norms are bytes
		return impacts
	}

	freqNormPairs := treeset.NewWith(ImpactComparator, c.otherFreqNormPairs.Values()...)
	for i := range impacts {
		impact := impacts[i]
		c.add(impact, freqNormPairs)
	}

	items := freqNormPairs.Values()
	impacts = impacts[:0]
	for i := range items {
		v := items[i].(*Impact)
		impacts = append(impacts, v)
	}

	return impacts
}
