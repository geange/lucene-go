package search

// BulkScorer This class is used to score a range of documents at once, and is returned by Weight.bulkScorer. Only queries that have a more optimized means of scoring across a range of documents need to override this. Otherwise, a default implementation is wrapped around the Scorer returned by Weight.scorer.
type BulkScorer interface {
}
