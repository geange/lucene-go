package search

import (
	"github.com/emirpasic/gods/queues/priorityqueue"
)

// TopDocsCollector
// A base class for all collectors that return a TopDocs output. This collector allows easy extension
// by providing a single constructor which accepts a PriorityQueue as well as protected members for
// that priority queue and a counter of the number of total hits. Extending classes can override any
// of the methods to provide their own implementation, as well as avoid the use of the priority queue
// entirely by passing null to TopDocsCollector(PriorityQueue). In that case however, you might want
// to consider overriding all methods, in order to avoid a NullPointerException.
type TopDocsCollector interface {
	Collector

	// PopulateResults
	// Populates the results array with the ScoreDoc instances.
	// This can be overridden in case a different ScoreDoc type should be returned.
	PopulateResults(results []*ScoreDoc, howMany int)

	// NewTopDocs
	// Returns a TopDocs instance containing the given results.
	// If results is null it means there are no results to return, either because
	// there were 0 calls to collect() or because the arguments to topDocs were invalid.
	NewTopDocs(results []*ScoreDoc, howMany int) (*TopDocs, error)

	// GetTotalHits
	// The total number of documents that matched this query.
	GetTotalHits() int

	// TopDocsSize
	// The number of valid PQ entries
	TopDocsSize() int

	// TopDocs
	// Returns the top docs that were collected by this collector.
	TopDocs() (*TopDocs, error)

	// TopDocsFrom
	// Returns the documents in the range [start .. pq.size()) that were collected by this collector.
	// Note that if start >= pq.size(), an empty TopDocs is returned. This method is convenient to
	// call if the application always asks for the last results, starting from the last 'page'.
	// NOTE: you cannot call this method more than once for each search execution.
	// If you need to call it more than once, passing each time a different start,
	// you should call topDocs() and work with the returned TopDocs object,
	// which will contain all the results this search execution collected.
	TopDocsFrom(start int) (*TopDocs, error)

	// TopDocsRange
	// Returns the documents in the range [start .. start+howMany) that were collected by this collector.
	// Note that if start >= pq.size(), an empty TopDocs is returned, and if pq.size() - start < howMany,
	// then only the available documents in [start .. pq.size()) are returned.
	// This method is useful to call in case pagination of search results is allowed by the search application,
	// as well as it attempts to optimize the memory used by allocating only as much as requested by howMany.
	// NOTE: you cannot call this method more than once for each search execution.
	// If you need to call it more than once, passing each time a different range,
	// you should call topDocs() and work with the returned TopDocs object,
	// which will contain all the results this search execution collected.
	TopDocsRange(start, howMany int) (*TopDocs, error)
}

type TopDocsCollectorDefault struct {
	pq *priorityqueue.Queue
}
