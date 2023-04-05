package search

// TopScoreDocCollector
// A Collector implementation that collects the top-scoring hits,
// returning them as a TopDocs. This is used by IndexSearcher to implement TopDocs-based search.
// Hits are sorted by score descending and then (when the scores are tied) docID ascending.
// When you create an instance of this collector you should know in advance whether documents
// are going to be collected in doc Id order or not.
//
// NOTE: The values Float.NaN and Float.NEGATIVE_INFINITY are not valid scores. This collector
// will not properly collect hits with such scores.
type TopScoreDocCollector struct {
}
