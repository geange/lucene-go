package search

// TwoPhaseIterator Returned by Scorer.twoPhaseIterator() to expose an approximation of a DocIdSetIterator. When the approximation()'s DocIdSetIterator.nextDoc() or DocIdSetIterator.advance(int) return, matches() needs to be checked in order to know whether the returned doc ID actually matches.
type TwoPhaseIterator interface {
}
