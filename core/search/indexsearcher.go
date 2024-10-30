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

var _ index.IndexSearcher = &IndexSearcher{}

// IndexSearcher
// Implements search over a single Reader.
// Applications usually need only call the inherited search(Query, int) method. For performance reasons, if your
// index is unchanging, you should share a single IndexSearcher instance across multiple searches instead of
// creating a new one per-search. If your index has changed and you wish to see the changes reflected in searching,
// you should use DirectoryReader.openIfChanged(DirectoryReader) to obtain a new reader and then create a new
// IndexSearcher from that. Also, for low-latency turnaround it's best to use a near-real-time reader
// (DirectoryReader.open(IndexWriter)). Once you have a new Reader, it's relatively cheap to create a
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
	leafContexts  []index.LeafReaderContext

	// used with executor - each slice holds a set of leafs executed within one thread
	leafSlices []index.LeafSlice

	// the default Similarity
	similarity index.Similarity

	queryCache         index.QueryCache
	queryCachingPolicy index.QueryCachingPolicy
}

func (r *IndexSearcher) GetQueryCache() (index.QueryCache, error) {
	return nil, errors.New("unsupported queryCache")
}

func (r *IndexSearcher) SetQueryCachingPolicy(queryCachingPolicy index.QueryCachingPolicy) {
	return
}

func (r *IndexSearcher) GetQueryCachingPolicy() (index.QueryCachingPolicy, error) {
	return nil, errors.New("unsupported queryCachingPolicy")
}

func (r *IndexSearcher) Slices(leaves []index.LeafReaderContext) []index.LeafSlice {
	slices := make([]index.LeafSlice, 0, len(leaves))
	for _, leaf := range leaves {
		item := index.LeafSlice{Leaves: []index.LeafReaderContext{leaf}}
		slices = append(slices, item)
	}
	return slices
}

func (r *IndexSearcher) Doc(ctx context.Context, docID int) (*document.Document, error) {
	return r.reader.Document(ctx, docID)
}

func (r *IndexSearcher) DocWithVisitor(ctx context.Context, docId int, fieldVisitor document.StoredFieldVisitor) error {
	return r.reader.DocumentWithVisitor(ctx, docId, fieldVisitor)
}

func (r *IndexSearcher) DocLimitFields(ctx context.Context, docId int, fieldsToLoad []string) (*document.Document, error) {
	return r.reader.DocumentWithFields(ctx, docId, fieldsToLoad)
}

func (r *IndexSearcher) Count(query index.Query) (int, error) {
	query, err := r.Rewrite(query)
	if err != nil {
		return 0, err
	}

	for {
		// remove wrappers that don't matter for counts
		if csQuery, ok := query.(*ConstantScoreQuery); ok {
			query = csQuery.GetQuery()
		} else {
			break
		}
	}

	// some counts can be computed in constant time
	if _, ok := query.(*MatchAllDocsQuery); ok {
		return r.reader.NumDocs(), nil
	}

	if termQuery, ok := query.(*TermQuery); ok {
		term := termQuery.GetTerm()
		count := 0

		leaves, err := r.reader.Leaves()
		if err != nil {
			return 0, err
		}

		for _, leaf := range leaves {
			docFreq, err := leaf.LeafReader().DocFreq(nil, term)
			if err != nil {
				return 0, err
			}
			count += docFreq
		}
		return count, nil
	}

	v, err := r.SearchByCollectorManager(nil, query, &collectorManager{})
	if err != nil {
		return 0, err
	}
	return v.(int), nil
}

var _ CollectorManager = &collectorManager{}

type collectorManager struct {
}

func (c *collectorManager) NewCollector() (index.Collector, error) {
	return NewTotalHitCountCollector(), nil
}

func (c *collectorManager) Reduce(collectors []index.Collector) (any, error) {
	total := 0
	for _, collector := range collectors {
		total += collector.(*TotalHitCountCollector).GetTotalHits()
	}
	return total, nil
}

func (r *IndexSearcher) GetSlices() []index.LeafSlice {
	return r.leafSlices
}

func NewIndexSearcher(r index.IndexReader) (index.IndexSearcher, error) {
	ctx, err := r.GetContext()
	if err != nil {
		return nil, err
	}
	return newIndexSearcher(ctx)
}

func newIndexSearcher(readerContext index.IndexReaderContext) (*IndexSearcher, error) {
	leaves, err := readerContext.Leaves()
	if err != nil {
		return nil, err
	}

	similarity, err := NewBM25Similarity()
	if err != nil {
		return nil, err
	}

	return &IndexSearcher{
		reader:             readerContext.Reader(),
		readerContext:      readerContext,
		leafContexts:       leaves,
		leafSlices:         nil,
		similarity:         similarity,
		queryCache:         nil,
		queryCachingPolicy: nil,
	}, nil
}

func (r *IndexSearcher) GetTopReaderContext() index.IndexReaderContext {
	return r.readerContext
}

func (r *IndexSearcher) GetIndexReader() index.IndexReader {
	return r.reader
}

func (r *IndexSearcher) SetSimilarity(similarity index.Similarity) {
	r.similarity = similarity
}

func (r *IndexSearcher) SetQueryCache(queryCache index.QueryCache) {
	r.queryCache = queryCache
}

func (r *IndexSearcher) Search(ctx context.Context, query index.Query, results index.Collector) error {
	query, err := r.Rewrite(query)
	if err != nil {
		return err
	}

	weight, err := r.CreateWeight(query, results.ScoreMode(), 1)
	if err != nil {
		return err
	}

	return r.SearchLeaves(ctx, r.leafContexts, weight, results)
}

// SearchAfter
// Finds the top n hits for query where all results are after a previous result (after).
// By passing the bottom result from a previous page as after, this method can be used for
// efficient 'deep-paging' across potentially large result sets.
// Throws: BooleanQuery.TooManyClauses – If a query would exceed BooleanQuery.getMaxClauseCount() clauses.
func (r *IndexSearcher) SearchAfter(ctx context.Context, after index.ScoreDoc, query index.Query, numHits int) (index.TopDocs, error) {
	limit := max(1, r.reader.MaxDoc())
	if after != nil && after.GetDoc() >= limit {
		return nil, errors.New("after.doc exceeds the number of documents in the reader")
	}

	cappedNumHits := min(numHits, limit)

	manager := &searchAfterCollectorManager{
		cappedNumHits: cappedNumHits,
		after:         after,
	}

	if len(r.leafSlices) <= 1 {
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

	v, err := r.SearchByCollectorManager(nil, query, manager)
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
func (r *IndexSearcher) SearchByCollectorManager(ctx context.Context,
	query index.Query, collectorManager CollectorManager) (any, error) {

	if len(r.leafSlices) <= 1 {
		collector, err := collectorManager.NewCollector()
		if err != nil {
			return nil, err
		}
		if err := r.SearchCollector(nil, query, collector); err != nil {
			return nil, err
		}
		return collectorManager.Reduce([]index.Collector{collector.(TopScoreDocCollector)})
	}

	// TODO: fix it
	collectors := make([]index.Collector, 0, len(r.leafSlices))
	var scoreMode *index.ScoreMode
	for i := 0; i < len(r.leafSlices); i++ {
		collector, err := collectorManager.NewCollector()
		if err != nil {
			return nil, err
		}
		collectors = append(collectors, collector)

		if scoreMode == nil {
			mode := collector.ScoreMode()
			scoreMode = &mode
		} else {
			mode := collector.ScoreMode()
			if mode != *scoreMode {
				return nil, errors.New("CollectorManager does not always produce collectors with the same score mode")
			}
		}
	}

	if scoreMode == nil {
		// no segments
		scoreMode = &COMPLETE
	}

	query, err := r.Rewrite(query)
	if err != nil {
		return nil, err
	}

	weight, err := r.CreateWeight(query, *scoreMode, 1)
	if err != nil {
		return nil, err
	}

	for i, item := range r.leafSlices {
		leaves := item.Leaves

		collector := collectors[i]

		// TODO: try use WaitGroup
		err := r.SearchLeaves(ctx, leaves, weight, collector)
		if err != nil {
			return nil, err
		}
	}

	return collectorManager.Reduce(collectors)
}

func (r *IndexSearcher) SearchTopN(ctx context.Context, query index.Query, n int) (index.TopDocs, error) {
	return r.SearchAfter(ctx, nil, query, n)
}

func (r *IndexSearcher) SearchCollector(ctx context.Context, query index.Query, results index.Collector) error {
	query, err := r.Rewrite(query)
	if err != nil {
		return err
	}

	weight, err := r.CreateWeight(query, results.ScoreMode(), 1)
	if err != nil {
		return err
	}
	return r.SearchLeaves(ctx, r.leafContexts, weight, results)
}

func (r *IndexSearcher) SearchLeaves(ctx context.Context, leaves []index.LeafReaderContext, weight index.Weight, collector index.Collector) error {

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

func (r *IndexSearcher) CreateWeight(query index.Query, scoreMode index.ScoreMode, boost float64) (index.Weight, error) {
	//queryCache := r.queryCache
	weight, err := query.CreateWeight(r, scoreMode, boost)
	if err != nil {
		return nil, err
	}

	//if !scoreMode.NeedsScores() && queryCache != nil {
	//	weight = queryCache.DoCache(weight, r.queryCachingPolicy)
	//}
	return weight, nil
}

func (r *IndexSearcher) Rewrite(query index.Query) (index.Query, error) {
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
func (r *IndexSearcher) GetSimilarity() index.Similarity {
	return r.similarity
}

// CollectionStatistics
// Returns CollectionStatistics for a field, or null if the field does not exist (has no indexed terms) This can be overridden for example, to return a field's statistics across a distributed collection.
func (r *IndexSearcher) CollectionStatistics(field string) (types.CollectionStatistics, error) {
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
func (r *IndexSearcher) TermStatistics(term index.Term, docFreq, totalTermFreq int) (types.TermStatistics, error) {
	return types.NewTermStatistics(term.Bytes(), int64(docFreq), int64(totalTermFreq))
}

type Executor interface {
}
