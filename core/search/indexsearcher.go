package search

import (
	"context"
	"errors"
	"reflect"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

const (
	TOTAL_HITS_THRESHOLD = 1000
)

var _ index.IndexSearcher = &indexSearcher{}

// indexSearcher
// Implements search over a single Reader.
// Applications usually need only call the inherited search(Query, int) method. For performance reasons, if your
// index is unchanging, you should share a single indexSearcher instance across multiple searches instead of
// creating a new one per-search. If your index has changed and you wish to see the changes reflected in searching,
// you should use DirectoryReader.openIfChanged(DirectoryReader) to obtain a new reader and then create a new
// indexSearcher from that. Also, for low-latency turnaround it's best to use a near-real-time reader
// (DirectoryReader.open(IndexWriter)). Once you have a new Reader, it's relatively cheap to create a
// new indexSearcher from it.
//
// NOTE: The search and searchAfter methods are configured to only count top hits accurately up to 1,000 and may
// return a lower bound of the hit count if the hit count is greater than or equal to 1,000. On queries that match
// lots of documents, counting the number of hits may take much longer than computing the top hits so this
// trade-off allows to get some minimal information about the hit count without slowing down search too much.
// The TopDocs.scoreDocs array is always accurate however. If this behavior doesn't suit your needs, you should
// create collectors manually with either TopScoreDocCollector.create or TopFieldCollector.create and call
// search(Query, Collector).
//
// NOTE: indexSearcher instances are completely thread safe, meaning multiple threads can call any of its
// methods, concurrently. If your application requires external synchronization, you should not synchronize on
// the indexSearcher instance; use your own (non-Lucene) objects instead.
type indexSearcher struct {
	reader index.IndexReader

	// NOTE: these members might change in incompatible ways
	// in the next release
	readerContext index.IndexReaderContext
	leafContexts  []index.LeafReaderContext

	// used with executor - each slice holds a set of leafs executed within one thread
	leafSlices []index.LeafSlice

	// These are only used for multi-threaded search
	executor Executor

	// the default Similarity
	similarity index.Similarity

	queryCache         index.QueryCache
	queryCachingPolicy index.QueryCachingPolicy
}

func (r *indexSearcher) GetQueryCache() index.QueryCache {
	//TODO implement me
	panic("implement me")
}

func (r *indexSearcher) SetQueryCachingPolicy(queryCachingPolicy index.QueryCachingPolicy) {
	//TODO implement me
	panic("implement me")
}

func (r *indexSearcher) GetQueryCachingPolicy() index.QueryCachingPolicy {
	//TODO implement me
	panic("implement me")
}

func (r *indexSearcher) Slices(leaves []index.LeafReaderContext) []index.LeafSlice {
	//TODO implement me
	panic("implement me")
}

func (r *indexSearcher) Doc(docID int) (*document.Document, error) {
	//TODO implement me
	panic("implement me")
}

func (r *indexSearcher) DocWithVisitor(docID int, fieldVisitor document.StoredFieldVisitor) (*document.Document, error) {
	//TODO implement me
	panic("implement me")
}

func (r *indexSearcher) DocLimitFields(docID int, fieldsToLoad []string) (*document.Document, error) {
	//TODO implement me
	panic("implement me")
}

func (r *indexSearcher) Count(query index.Query) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (r *indexSearcher) GetSlices() []index.LeafSlice {
	//TODO implement me
	panic("implement me")
}

func NewIndexSearcher(r index.IndexReader) (index.IndexSearcher, error) {
	ctx, err := r.GetContext()
	if err != nil {
		return nil, err
	}
	return newIndexSearcher(ctx)
}

func newIndexSearcher(readerContext index.IndexReaderContext) (*indexSearcher, error) {
	leaves, err := readerContext.Leaves()
	if err != nil {
		return nil, err
	}

	similarity, err := NewBM25Similarity()
	if err != nil {
		return nil, err
	}

	return &indexSearcher{
		reader:             readerContext.Reader(),
		readerContext:      readerContext,
		leafContexts:       leaves,
		leafSlices:         nil,
		similarity:         similarity,
		queryCache:         nil,
		queryCachingPolicy: nil,
	}, nil
}

func (r *indexSearcher) GetTopReaderContext() index.IndexReaderContext {
	return r.readerContext
}

func (r *indexSearcher) GetIndexReader() index.IndexReader {
	return r.reader
}

func (r *indexSearcher) SetSimilarity(similarity index.Similarity) {
	r.similarity = similarity
}

func (r *indexSearcher) SetQueryCache(queryCache index.QueryCache) {
	r.queryCache = queryCache
}

func (r *indexSearcher) Search(query index.Query, results index.Collector) error {
	query, err := r.Rewrite(query)
	if err != nil {
		return err
	}

	weight, err := r.CreateWeight(query, results.ScoreMode(), 1)
	if err != nil {
		return err
	}

	return r.Search3(r.leafContexts, weight, results)
}

// SearchAfter
// Finds the top n hits for query where all results are after a previous result (after).
// By passing the bottom result from a previous page as after, this method can be used for
// efficient 'deep-paging' across potentially large result sets.
// Throws: BooleanQuery.TooManyClauses – If a query would exceed BooleanQuery.getMaxClauseCount() clauses.
func (r *indexSearcher) SearchAfter(after index.ScoreDoc, query index.Query, numHits int) (index.TopDocs, error) {
	limit := max(1, r.reader.MaxDoc())
	if after != nil && after.GetDoc() >= limit {
		return nil, errors.New("after.doc exceeds the number of documents in the reader")
	}

	cappedNumHits := min(numHits, limit)

	manager := &searchAfterCollectorManager{
		cappedNumHits: cappedNumHits,
		after:         after,
	}

	if r.executor == nil || len(r.leafSlices) <= 1 {
		hitsThresholdChecker, err := HitsThresholdCheckerCreate(max(TOTAL_HITS_THRESHOLD, numHits))
		if err != nil {
			return nil, err
		}
		manager.hitsThresholdChecker = hitsThresholdChecker
	} else {
		manager.minScoreAcc = NewMaxScoreAccumulator()
		hitsThresholdChecker, err := HitsThresholdCheckerCreateShared(max(TOTAL_HITS_THRESHOLD, numHits))
		if err != nil {
			return nil, err
		}
		manager.hitsThresholdChecker = hitsThresholdChecker
	}

	v, err := r.SearchByCollectorManager(query, manager)
	if err != nil {
		return nil, err
	}

	topDocs, ok := v.(index.TopDocs)
	if !ok {
		return nil, errors.New("object is not TopDocs")
	}

	return topDocs, nil
}

var _ CollectorManager = &searchAfterCollectorManager{}

type searchAfterCollectorManager struct {
	hitsThresholdChecker HitsThresholdChecker
	minScoreAcc          *MaxScoreAccumulator
	cappedNumHits        int
	after                index.ScoreDoc
}

func (s *searchAfterCollectorManager) NewCollector() (index.Collector, error) {
	return TopScoreDocCollectorCreate(s.cappedNumHits, s.after, s.hitsThresholdChecker, s.minScoreAcc)
}

func (s *searchAfterCollectorManager) Reduce(collectors []index.Collector) (any, error) {
	topDocs := make([]index.TopDocs, len(collectors))
	for i, collector := range collectors {
		docs, err := collector.(TopScoreDocCollector).TopDocs()
		if err != nil {
			return nil, err
		}
		topDocs[i] = docs
	}
	return MergeTopDocs(0, s.cappedNumHits, topDocs, true)
}

// SearchByCollectorManager
// Lower-level search API. Search all leaves using the given CollectorManager.
// In contrast to search(Query, Collector), this method will use the searcher's
// Executor in order to parallelize execution of the collection on the configured leafSlices.
// See Also: CollectorManager
// lucene.experimental
func (r *indexSearcher) SearchByCollectorManager(query index.Query, collectorManager CollectorManager) (any, error) {
	if r.executor == nil || len(r.leafSlices) <= 1 {
		collector, err := collectorManager.NewCollector()
		if err != nil {
			return nil, err
		}
		if err := r.SearchCollector(query, collector); err != nil {
			return nil, err
		}
		return collectorManager.Reduce([]index.Collector{collector.(TopScoreDocCollector)})
	}

	// TODO: fix it
	panic("")
}

func (r *indexSearcher) SearchTopN(query index.Query, n int) (index.TopDocs, error) {
	return r.SearchAfter(nil, query, n)
}

func (r *indexSearcher) SearchCollector(query index.Query, results index.Collector) error {
	query, err := r.Rewrite(query)
	if err != nil {
		return err
	}

	weight, err := r.CreateWeight(query, results.ScoreMode(), 1)
	if err != nil {
		return err
	}
	return r.Search3(r.leafContexts, weight, results)
}

func (r *indexSearcher) Search3(leaves []index.LeafReaderContext, weight index.Weight, collector index.Collector) error {

	for _, leaf := range leaves {
		leafCollector, err := collector.GetLeafCollector(context.TODO(), leaf)
		if err != nil {
			continue
		}

		scorer, err := weight.BulkScorer(leaf)
		if err != nil {
			return err
		}

		if scorer != nil {
			if err := scorer.Score(leafCollector, leaf.LeafReader().GetLiveDocs()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *indexSearcher) CreateWeight(query index.Query, scoreMode index.ScoreMode, boost float64) (index.Weight, error) {
	queryCache := r.queryCache
	weight, err := query.CreateWeight(r, scoreMode, boost)
	if err != nil {
		return nil, err
	}

	if !scoreMode.NeedsScores() && queryCache != nil {
		weight = queryCache.DoCache(weight, r.queryCachingPolicy)
	}
	return weight, nil
}

func (r *indexSearcher) Rewrite(query index.Query) (index.Query, error) {
	rewrittenQuery, err := query.Rewrite(r.reader)
	if err != nil {
		return nil, err
	}

	for !reflect.DeepEqual(rewrittenQuery, query) {
		query = rewrittenQuery
		rewrittenQuery, err = query.Rewrite(r.reader)
		if err != nil {
			return nil, err
		}
	}

	return query, nil
}

// GetSimilarity
// Expert: Get the Similarity to use to compute scores. This returns the Similarity
// that has been set through setSimilarity(Similarity) or the default Similarity if none has been set explicitly.
func (r *indexSearcher) GetSimilarity() index.Similarity {
	return r.similarity
}

// CollectionStatistics
// Returns CollectionStatistics for a field, or null if the field does not exist (has no indexed terms) This can be overridden for example, to return a field's statistics across a distributed collection.
func (r *indexSearcher) CollectionStatistics(field string) (types.CollectionStatistics, error) {
	docCount := 0
	sumTotalTermFreq := int64(0)
	sumDocFreq := int64(0)

	leaves, err := r.reader.Leaves()
	if err != nil {
		return nil, err
	}

	for _, leaf := range leaves {
		terms, err := leaf.LeafReader().Terms(field)
		if err != nil {
			return nil, err
		}
		if terms == nil {
			continue
		}

		count, err := terms.GetDocCount()
		if err != nil {
			return nil, err
		}
		docCount += count

		totalTermFreq, err := terms.GetSumTotalTermFreq()
		if err != nil {
			return nil, err
		}
		sumTotalTermFreq += totalTermFreq

		docFreq, err := terms.GetSumDocFreq()
		if err != nil {
			return nil, err
		}
		sumDocFreq += docFreq
	}

	if docCount == 0 {
		return nil, nil
	}

	return types.NewCollectionStatistics(field, int64(r.reader.MaxDoc()), int64(docCount), sumTotalTermFreq, sumDocFreq)
}

// TermStatistics
// Returns TermStatistics for a term.
// This can be overridden for example, to return a term's statistics across a distributed collection.
// Params:
// docFreq – The document frequency of the term. It must be greater or equal to 1.
// totalTermFreq – The total term frequency.
// Returns:
// A TermStatistics (never null).
func (r *indexSearcher) TermStatistics(term index.Term, docFreq, totalTermFreq int) (types.TermStatistics, error) {
	return types.NewTermStatistics(term.Bytes(), int64(docFreq), int64(totalTermFreq))
}

type Executor interface {
}
