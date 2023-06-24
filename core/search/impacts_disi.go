package search

import (
	"github.com/geange/lucene-go/core/index"
	"math"
)

var _ index.DocIdSetIterator = &ImpactsDISI{}

// ImpactsDISI DocIdSetIterator that skips non-competitive docs thanks to the indexed impacts.
// Call SetMinCompetitiveScore(float) in order to give this iterator the ability to skip
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

func NewImpactsDISI(in index.DocIdSetIterator, impactsSource index.ImpactsSource, scorer index.SimScorer) *ImpactsDISI {
	return &ImpactsDISI{
		in:             in,
		impactsSource:  impactsSource,
		maxScoreCache:  NewMaxScoreCache(impactsSource, scorer),
		globalMaxScore: scorer.Score(math.MaxFloat32, 1),
	}
}

func (d *ImpactsDISI) setMinCompetitiveScore(minCompetitiveScore float64) error {
	// assert minCompetitiveScore >= this.minCompetitiveScore;
	if d.minCompetitiveScore > d.minCompetitiveScore {
		d.minCompetitiveScore = minCompetitiveScore
		// force upTo and maxScore to be recomputed so that we will skip documents
		// if the current block of documents is not competitive - only if the min
		// competitive score actually increased
		d.upTo = -1
	}
	return nil
}

// Implement the contract of Scorer.advanceShallow(int) based on the wrapped ImpactsEnum.
// See Also: Scorer.advanceShallow(int)
func (d *ImpactsDISI) advanceShallow(target int) (int, error) {
	err := d.impactsSource.AdvanceShallow(target)
	if err != nil {
		return 0, err
	}

	impacts, err := d.impactsSource.GetImpacts()
	if err != nil {
		return 0, err
	}
	return impacts.GetDocIdUpTo(0), nil
}

// GetMaxScore
// Implement the contract of Scorer.GetMaxScore(int) based on the wrapped ImpactsEnum and Scorer.
// See Also: Scorer.GetMaxScore(int)
func (d *ImpactsDISI) GetMaxScore(upTo int) (float64, error) {
	level, err := d.maxScoreCache.GetLevel(upTo)
	if err != nil {
		return 0, err
	}
	if level == -1 {
		return d.globalMaxScore, nil
	}
	return d.maxScoreCache.GetMaxScoreForLevel(level)
}

func (d *ImpactsDISI) DocID() int {
	return d.in.DocID()
}

func (d *ImpactsDISI) NextDoc() (int, error) {
	return d.Advance(d.in.DocID() + 1)
}

func (d *ImpactsDISI) Advance(target int) (int, error) {
	target, err := d.advanceTarget(target)
	if err != nil {
		return 0, err
	}
	return d.in.Advance(target)
}

func (d *ImpactsDISI) advanceTarget(target int) (int, error) {
	if target <= d.upTo {
		// we are still in the current block, which is considered competitive
		// according to impacts, no skipping
		return target, nil
	}

	var err error
	d.upTo, err = d.advanceShallow(target)
	if err != nil {
		return 0, err
	}
	maxScore, err := d.maxScoreCache.GetMaxScoreForLevel(0)
	if err != nil {
		return 0, err
	}

	for {
		//assert upTo >= target;

		if maxScore >= d.minCompetitiveScore {
			return target, nil
		}

		if d.upTo == index.NO_MORE_DOCS {
			return index.NO_MORE_DOCS, nil
		}

		skipUpTo, err := d.maxScoreCache.GetSkipUpTo(d.minCompetitiveScore)
		if err != nil {
			return 0, err
		}
		if skipUpTo == -1 { // no further skipping
			target = d.upTo + 1
		} else if skipUpTo == index.NO_MORE_DOCS {
			return index.NO_MORE_DOCS, nil
		} else {
			target = skipUpTo + 1
		}
		d.upTo, err = d.advanceShallow(target)
		if err != nil {
			return 0, err
		}
		d.maxScore, err = d.maxScoreCache.GetMaxScoreForLevel(0)
		if err != nil {
			return 0, err
		}
	}
}

func (d *ImpactsDISI) SlowAdvance(target int) (int, error) {
	return d.Advance(target)
}

func (d *ImpactsDISI) Cost() int64 {
	return d.in.Cost()
}
