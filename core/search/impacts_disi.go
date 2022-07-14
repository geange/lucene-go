package search

// ImpactsDISI DocIdSetIterator that skips non-competitive docs thanks to the indexed impacts.
// Call setMinCompetitiveScore(float) in order to give this iterator the ability to skip
// low-scoring documents.
type ImpactsDISI struct {
}

func (d *ImpactsDISI) setMinCompetitiveScore(score float64) error {
	panic("")
}

func (d *ImpactsDISI) getMaxScore(to int) (float64, error) {
	panic("")
}
