package search

// QueryCache A cache for queries.
// See Also: LRUQueryCache
type QueryCache interface {

	// DoCache Return a wrapper around the provided weight that will cache matching docs per-segment accordingly to
	// the given policy. NOTE: The returned weight will only be equivalent if scores are not needed.
	// See Also: Collector.scoreMode()
	DoCache(weight Weight, policy QueryCachingPolicy) Weight
}
