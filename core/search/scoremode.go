package search

import (
	"github.com/geange/lucene-go/core/interface/index"
)

var (
	// COMPLETE
	// Produced scorers will allow visiting all matches and get their score.
	COMPLETE = index.NewScoreMode(true, true)

	// COMPLETE_NO_SCORES
	// Produced scorers will allow visiting all matches but scores won't be available.
	COMPLETE_NO_SCORES = index.NewScoreMode(true, false)

	// TOP_SCORES
	// Produced scorers will optionally allow skipping over non-competitive hits using the Scorer.SetMinCompetitiveScore(float) API.
	TOP_SCORES = index.NewScoreMode(false, true)

	// TOP_DOCS
	// ScoreMode for top field collectors that can provide their own iterators, to optionally allow to skip for non-competitive docs
	TOP_DOCS = index.NewScoreMode(false, false)

	// TOP_DOCS_WITH_SCORES
	// ScoreMode for top field collectors that can provide their own iterators, to optionally allow to skip for non-competitive docs. This mode is used when there is a secondary sort by _score.
	TOP_DOCS_WITH_SCORES = index.NewScoreMode(false, true)
)
