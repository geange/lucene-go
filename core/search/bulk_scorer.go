package search

import "github.com/bits-and-blooms/bitset"

// BulkScorer This class is used to score a range of documents at once, and is returned by Weight.bulkScorer.
// Only queries that have a more optimized means of scoring across a range of documents need to override this.
// Otherwise, a default implementation is wrapped around the Scorer returned by Weight.scorer.
type BulkScorer interface {
	// Score Scores and collects all matching documents.
	// Params: 	collector – The collector to which all matching documents are passed.
	//			acceptDocs – Bits that represents the allowed documents to match, or null if they are all allowed to match.
	Score(collector LeafCollector, acceptDocs *bitset.BitSet) error
}
