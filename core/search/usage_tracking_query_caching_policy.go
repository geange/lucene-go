package search

// UsageTrackingQueryCachingPolicy A QueryCachingPolicy that tracks usage statistics of recently-used
// filters in order to decide on which filters are worth caching.
type UsageTrackingQueryCachingPolicy struct {
}

func (u *UsageTrackingQueryCachingPolicy) OnUse(query Query) {
	//TODO implement me
	panic("implement me")
}

func (u *UsageTrackingQueryCachingPolicy) ShouldCache(query Query) (bool, error) {
	//TODO implement me
	panic("implement me")
}
