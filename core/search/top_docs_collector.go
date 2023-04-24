package search

import (
	"github.com/geange/lucene-go/core/util/structure"
	"github.com/geange/lucene-go/math"
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
	PopulateResults(results []ScoreDoc, howMany int) error

	// NewTopDocs
	// Returns a TopDocs instance containing the given results.
	// If results is null it means there are no results to return, either because
	// there were 0 calls to collect() or because the arguments to topDocs were invalid.
	NewTopDocs(results []ScoreDoc, howMany int) (TopDocs, error)

	// GetTotalHits
	// The total number of documents that matched this query.
	GetTotalHits() int

	// TopDocsSize
	// The number of valid PQ entries
	TopDocsSize() int

	// TopDocs
	// Returns the top docs that were collected by this collector.
	TopDocs() (TopDocs, error)

	// TopDocsFrom
	// Returns the documents in the range [start .. pq.size()) that were collected by this collector.
	// Note that if start >= pq.size(), an empty TopDocs is returned. This method is convenient to
	// call if the application always asks for the last results, starting from the last 'page'.
	// NOTE: you cannot call this method more than once for each search execution.
	// If you need to call it more than once, passing each time a different start,
	// you should call topDocs() and work with the returned TopDocs object,
	// which will contain all the results this search execution collected.
	TopDocsFrom(start int) (TopDocs, error)

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
	TopDocsRange(start, howMany int) (TopDocs, error)
}

var EMPTY_TOPDOCS = &TopDocsDefault{
	totalHits: NewTotalHits(0, EQUAL_TO),
	scoreDocs: make([]ScoreDoc, 0),
}

type TopDocsCollectorDefault[T ScoreDoc] struct {
	pq                *structure.PriorityQueue[T]
	totalHits         int
	totalHitsRelation TotalHitsRelation
}

func newTopDocsCollectorDefault[T ScoreDoc](pq *structure.PriorityQueue[T]) *TopDocsCollectorDefault[T] {
	return &TopDocsCollectorDefault[T]{pq: pq}
}

func (t *TopDocsCollectorDefault[T]) PopulateResults(results []ScoreDoc, howMany int) error {
	for i := howMany - 1; i >= 0; i-- {
		n, err := t.pq.Pop()
		if err != nil {
			return err
		}
		results[i] = n
	}
	return nil
}

func (t *TopDocsCollectorDefault[T]) NewTopDocs(results []ScoreDoc, howMany int) (TopDocs, error) {
	if len(results) == 0 {
		return EMPTY_TOPDOCS, nil
	}
	return NewTopDocs(NewTotalHits(int64(t.totalHits), t.totalHitsRelation), results), nil
}

func (t *TopDocsCollectorDefault[T]) GetTotalHits() int {
	return t.totalHits
}

func (t *TopDocsCollectorDefault[T]) TopDocsSize() int {
	// In case pq was populated with sentinel values, there might be less
	// results than pq.size(). Therefore return all results until either
	// pq.size() or totalHits.
	if t.totalHits < t.pq.Size() {
		return t.totalHits
	}
	return t.pq.Size()
}

func (t *TopDocsCollectorDefault[T]) TopDocs() (TopDocs, error) {
	return t.TopDocsRange(0, t.TopDocsSize())
}

func (t *TopDocsCollectorDefault[T]) TopDocsFrom(start int) (TopDocs, error) {
	return t.TopDocsRange(start, t.TopDocsSize())
}

func (t *TopDocsCollectorDefault[T]) TopDocsRange(start, howMany int) (TopDocs, error) {
	// In case pq was populated with sentinel values, there might be less
	// results than pq.size(). Therefore return all results until either
	// pq.size() or totalHits.
	size := t.TopDocsSize()

	// Don't bother to throw an exception, just return an empty TopDocs in case
	// the parameters are invalid or out of range.
	// TODO: shouldn't we throw IAE if apps give bad params here so they dont
	// have sneaky silent bugs?
	if start < 0 || start >= size || howMany <= 0 {
		return t.NewTopDocs(nil, start)
	}

	// We know that start < pqsize, so just fix howMany.
	howMany = math.Min(size-start, howMany)
	results := make([]ScoreDoc, howMany)

	// pq's pop() returns the 'least' element in the queue, therefore need
	// to discard the first ones, until we reach the requested range.
	// Note that this loop will usually not be executed, since the common usage
	// should be that the caller asks for the last howMany results. However it's
	// needed here for completeness.
	for i := t.pq.Size() - start - howMany; i > 0; i-- {
		t.pq.Pop()
	}

	// Get the requested results from pq.
	if err := t.PopulateResults(results, howMany); err != nil {
		return nil, err
	}

	return t.NewTopDocs(results, start)
}
