package search

import "github.com/geange/lucene-go/core/index"

var _ Scorer = &WANDScorer{}

// WANDScorer
// This implements the WAND (Weak AND) algorithm for dynamic pruning described in "Efficient
// Query Evaluation using a Two-Level Retrieval Process" by Broder, Carmel, Herscovici, Soffer and Zien.
// Enhanced with techniques described in "Faster Top-k Document Retrieval Using Block-Max Indexes"
// by Ding and Suel. For scoreMode == ScoreMode.TOP_SCORES, this scorer maintains a feedback loop
// with the collector in order to know at any time the minimum score that is required in order for
// a hit to be competitive.
//
// The implementation supports both minCompetitiveScore by enforce that ∑ max_score >= minCompetitiveScore,
// and minShouldMatch by enforcing freq >= minShouldMatch. It keeps sub scorers in 3 different places: - tail: a heap that contains scorers that are behind the desired doc ID. These scorers are ordered by cost so that we can advance the least costly ones first. - lead: a linked list of scorer that are positioned on the desired doc ID - head: a heap that contains scorers which are beyond the desired doc ID, ordered by doc ID in order to move quickly to the next candidate.
// When scoreMode == ScoreMode.TOP_SCORES, it leverages the max score from each scorer in order to know when it may call DocIdSetIterator.advance rather than DocIdSetIterator.nextDoc to move to the next competitive hit. When scoreMode != ScoreMode.TOP_SCORES, block-max scoring related logic is skipped. Finding the next match consists of first setting the desired doc ID to the least entry in 'head', and then advance 'tail' until there is a match, by meeting the configured freq >= minShouldMatch and / or ∑ max_score >= minCompetitiveScore requirements.
type WANDScorer struct {
	*ScorerDefault

	scalingFactor       int
	minCompetitiveScore int64

	// TODO
}

func newWANDScorer(weight Weight, scorers []Scorer, minShouldMatch int, scoreMode *ScoreMode) (*WANDScorer, error) {
	panic("")
}

func (w *WANDScorer) Score() (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (w *WANDScorer) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (w *WANDScorer) Iterator() index.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}

func (w *WANDScorer) GetMaxScore(upTo int) (float64, error) {
	//TODO implement me
	panic("implement me")
}

// DisiWrapper
// Wrapper used in DisiPriorityQueue.
// lucene.internal
type DisiWrapper struct {
	iterator  index.DocIdSetIterator
	scorer    Scorer
	cost      int64
	matchCost float64      // the match cost for two-phase iterators, 0 otherwise
	doc       int          // the current doc, used for comparison
	next      *DisiWrapper // reference to a next element, see #topList

	// An approximation of the iterator, or the iterator itself if it does not
	// support two-phase iteration
	approximation index.DocIdSetIterator

	// A two-phase view of the iterator, or null if the iterator does not support
	// two-phase iteration
	twoPhaseView TwoPhaseIterator

	// For WANDScorer
	maxScore int64
}
