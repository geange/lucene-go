package search

import "github.com/geange/lucene-go/core/interface/search"

var _ search.QueryCache = &LRUQueryCache{}

// LRUQueryCache
// A QueryCache that evicts queries using a LRU (least-recently-used) eviction policy in order to remain
// under a given maximum size and number of bytes used. This class is thread-safe. Note that query eviction
// runs in linear time with the total number of segments that have cache entries so this cache works best
// with caching policies that only cache on "large" segments, and it is advised to not share this cache
// across too many indices. A default query cache and policy instance is used in IndexSearcher. If you want
// to replace those defaults it is typically done like this:
//
// ```
//
//	maxNumberOfCachedQueries := 256
//
// ```
//
//	final int maxNumberOfCachedQueries = 256;
//	final long maxRamBytesUsed = 50 * 1024L * 1024L; // 50MB
//	// these cache and policy instances can be shared across several queries and readers
//	// it is fine to eg. store them into static variables
//	final QueryCache queryCache = new LRUQueryCache(maxNumberOfCachedQueries, maxRamBytesUsed);
//	final QueryCachingPolicy defaultCachingPolicy = new UsageTrackingQueryCachingPolicy();
//	indexSearcher.setQueryCache(queryCache);
//	indexSearcher.setQueryCachingPolicy(defaultCachingPolicy);
//
// This cache exposes some global statistics (hit count, miss count, number of cache entries, total number of DocIdSets that have ever been cached, number of evicted entries). In case you would like to have more fine-grained statistics, such as per-index or per-query-class statistics, it is possible to override various callbacks: onHit, onMiss, onQueryCache, onQueryEviction, onDocIdSetCache, onDocIdSetEviction and onClear. It is better to not perform heavy computations in these methods though since they are called synchronously and under a lock.
// See Also: QueryCachingPolicy
// lucene.experimental
type LRUQueryCache struct {
}

func (c *LRUQueryCache) DoCache(weight search.Weight, policy search.QueryCachingPolicy) search.Weight {
	//TODO implement me
	panic("implement me")
}
