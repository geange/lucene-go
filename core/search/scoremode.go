package search

// ScoreMode Different modes of search.
type ScoreMode struct {
	isExhaustive bool
	needsScores  bool
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

func (r *ScoreMode) Equal(mode *ScoreMode) bool {
	return r.isExhaustive == mode.isExhaustive && r.needsScores == mode.needsScores
}

var (
	// COMPLETE Produced scorers will allow visiting all matches and get their score.
	COMPLETE = &ScoreMode{true, true}

	// COMPLETE_NO_SCORES Produced scorers will allow visiting all matches but scores won't be available.
	COMPLETE_NO_SCORES = &ScoreMode{true, false}

	// TOP_SCORES Produced scorers will optionally allow skipping over non-competitive hits using the Scorer.SetMinCompetitiveScore(float) API.
	TOP_SCORES = &ScoreMode{false, true}

	// TOP_DOCS ScoreMode for top field collectors that can provide their own iterators, to optionally allow to skip for non-competitive docs
	TOP_DOCS = &ScoreMode{false, false}

	// TOP_DOCS_WITH_SCORES ScoreMode for top field collectors that can provide their own iterators, to optionally allow to skip for non-competitive docs. This mode is used when there is a secondary sort by _score.
	TOP_DOCS_WITH_SCORES = &ScoreMode{false, true}
)
