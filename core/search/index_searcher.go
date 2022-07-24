package search

import (
	"github.com/geange/lucene-go/core/index"
	"reflect"
)

// IndexSearcher Implements search over a single IndexReader.
// Applications usually need only call the inherited search(Query, int) method. For performance reasons, if your
// index is unchanging, you should share a single IndexSearcher instance across multiple searches instead of
// creating a new one per-search. If your index has changed and you wish to see the changes reflected in searching,
// you should use DirectoryReader.openIfChanged(DirectoryReader) to obtain a new reader and then create a new
// IndexSearcher from that. Also, for low-latency turnaround it's best to use a near-real-time reader
// (DirectoryReader.open(IndexWriter)). Once you have a new IndexReader, it's relatively cheap to create a
// new IndexSearcher from it.
//
// NOTE: The search and searchAfter methods are configured to only count top hits accurately up to 1,000 and may
// return a lower bound of the hit count if the hit count is greater than or equal to 1,000. On queries that match
// lots of documents, counting the number of hits may take much longer than computing the top hits so this
// trade-off allows to get some minimal information about the hit count without slowing down search too much.
// The TopDocs.scoreDocs array is always accurate however. If this behavior doesn't suit your needs, you should
// create collectors manually with either TopScoreDocCollector.create or TopFieldCollector.create and call
// search(Query, Collector).
//
// NOTE: IndexSearcher instances are completely thread safe, meaning multiple threads can call any of its
// methods, concurrently. If your application requires external synchronization, you should not synchronize on
// the IndexSearcher instance; use your own (non-Lucene) objects instead.
type IndexSearcher struct {
	reader index.IndexReader

	// NOTE: these members might change in incompatible ways
	// in the next release
	readerContext index.IndexReaderContext
	leafContexts  []*index.LeafReaderContext

	// used with executor - each slice holds a set of leafs executed within one thread
	leafSlices []LeafSlice

	// These are only used for multi-threaded search
	//executor

	// the default Similarity
	similarity Similarity

	queryCache         QueryCache
	queryCachingPolicy QueryCachingPolicy
}

func NewIndexSearcher(r index.IndexReader) *IndexSearcher {
	return newIndexSearcher(r.GetContext())
}

func newIndexSearcher(context index.IndexReaderContext) *IndexSearcher {
	leaves, err := context.Leaves()
	if err != nil {
		return nil
	}

	return &IndexSearcher{
		reader:             context.Reader(),
		readerContext:      context,
		leafContexts:       leaves,
		leafSlices:         nil,
		similarity:         nil,
		queryCache:         nil,
		queryCachingPolicy: nil,
	}
}

func (r *IndexSearcher) GetTopReaderContext() index.IndexReaderContext {
	return r.readerContext
}

func (r *IndexSearcher) SetSimilarity(similarity Similarity) {
	r.similarity = similarity
}

func (r *IndexSearcher) SetQueryCache(queryCache QueryCache) {
	r.queryCache = queryCache
}

func (r *IndexSearcher) Search(query Query, results Collector) error {
	var err error
	query, err = r.Rewrite(query)
	if err != nil {
		return err
	}

	weight, err := r.createWeight(query, results.ScoreMode(), 1)
	if err != nil {
		return err
	}

	return r.Search3(r.leafContexts, weight, results)
}

func (r *IndexSearcher) Search3(leaves []*index.LeafReaderContext, weight Weight, collector Collector) error {

	for _, ctx := range leaves {
		leafCollector, err := collector.GetLeafCollector(ctx)
		if err != nil {
			continue
		}

		scorer, err := weight.BulkScorer(ctx)
		if err != nil {
			return err
		}
		if scorer != nil {
			err := scorer.Score(leafCollector, ctx.LeafReader().GetLiveDocs())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *IndexSearcher) createWeight(query Query, scoreMode *ScoreMode, boost float64) (Weight, error) {
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

func (r *IndexSearcher) Rewrite(query Query) (Query, error) {
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

type LeafSlice struct {
	Leaves []index.LeafReaderContext
}
