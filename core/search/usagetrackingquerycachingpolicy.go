package search

import (
	"github.com/geange/lucene-go/core/interface/index"
)

// UsageTrackingQueryCachingPolicy A QueryCachingPolicy that tracks usage statistics of recently-used
// filters in order to decide on which filters are worth caching.
type UsageTrackingQueryCachingPolicy struct {
}

func (u *UsageTrackingQueryCachingPolicy) OnUse(query index.Query) {
	if shouldNeverCache(query) {
		return
	}

	panic("")
}

func shouldNeverCache(query index.Query) bool {
	switch q := query.(type) {
	case *TermQuery:
		// We do not bother caching term queries since they are already plenty fast.
		return true
	case *DocValuesFieldExistsQuery:
		// We do not bother caching DocValuesFieldExistsQuery queries since they are already plenty fast.
		return true
	case *MatchAllDocsQuery:
		// MatchAllDocsQuery has an iterator that is faster than what a bit set could do.
		return true
	case *MatchNoDocsQuery:
		// For the below queries, it's cheap to notice they cannot match any docs so
		// we do not bother caching them.
		return true
	case *BooleanQuery:
		if len(q.Clauses()) == 0 {
			return true
		}
	case *DisjunctionMaxQuery:
		if len(q.GetDisjuncts()) == 0 {
			return true
		}
	}
	return false
}

func (u *UsageTrackingQueryCachingPolicy) ShouldCache(query index.Query) (bool, error) {
	if shouldNeverCache(query) {
		return false, nil
	}

	panic("")
}
