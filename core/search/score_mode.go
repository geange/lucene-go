package search

// ScoreMode Different modes of search.
type ScoreMode struct {
	needsScores  bool
	isExhaustive bool
}

// NeedsScores Whether this ScoreMode needs to compute scores.
func (r *ScoreMode) NeedsScores() bool {
	return r.needsScores
}

// IsExhaustive Returns true if for this ScoreMode it is necessary to process all documents, or false if
// is enough to go through top documents only.
func (r *ScoreMode) IsExhaustive() bool {
	return r.isExhaustive
}
