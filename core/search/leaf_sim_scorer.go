package search

import (
	"github.com/geange/lucene-go/core/index"
)

type LeafSimScorer struct {
	scorer index.SimScorer
	norms  index.NumericDocValues
}

func NewLeafSimScorer(scorer index.SimScorer, reader index.LeafReader,
	field string, needsScores bool) (*LeafSimScorer, error) {
	leafSimScorer := &LeafSimScorer{scorer: scorer}
	var err error
	if needsScores {
		leafSimScorer.norms, err = reader.GetNormValues(field)
		if err != nil {
			return nil, err
		}
	}
	return leafSimScorer, nil
}

func (r *LeafSimScorer) GetSimScorer() index.SimScorer {
	return r.scorer
}

func (r *LeafSimScorer) Score(doc int, freq float64) (float64, error) {
	value, err := r.getNormValue(doc)
	if err != nil {
		return 0, err
	}

	return r.scorer.Score(freq, value), nil
}

func (r *LeafSimScorer) getNormValue(doc int) (int64, error) {
	if r.norms != nil {
		_, err := r.norms.AdvanceExact(doc)
		if err != nil {
			return 0, err
		}
		return r.norms.LongValue()
	}
	return 1, nil
}
