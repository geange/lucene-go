package search

import (
	"github.com/geange/lucene-go/core/index"
	index2 "github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

type LeafSimScorer struct {
	scorer index.SimScorer
	norms  index2.NumericDocValues
}

// NewLeafSimScorer
// org.apache.lucene.search.similarities.Similarity.SimScorer on a specific LeafReader.
func NewLeafSimScorer(scorer index.SimScorer, reader index2.LeafReader,
	field string, needsScores bool) (*LeafSimScorer, error) {
	leafSimScorer := &LeafSimScorer{scorer: scorer}
	if needsScores {
		norms, err := reader.GetNormValues(field)
		if err != nil {
			return nil, err
		}
		leafSimScorer.norms = norms
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

// Explain
// the score for the provided document assuming the given term document frequency.
// This method must be called on non-decreasing sequences of doc ids.
// See Also:
// org.apache.lucene.search.similarities.Similarity.SimScorer.explain(Explanation, long)
func (r *LeafSimScorer) Explain(doc int, freqExp *types.Explanation) (*types.Explanation, error) {
	normValue, err := r.getNormValue(doc)
	if err != nil {
		return nil, err
	}
	return r.scorer.Explain(freqExp, normValue)
}

func (r *LeafSimScorer) getNormValue(doc int) (int64, error) {
	if r.norms != nil {
		if _, err := r.norms.AdvanceExact(doc); err != nil {
			return 0, err
		}
		return r.norms.LongValue()
	}
	return 1, nil
}
