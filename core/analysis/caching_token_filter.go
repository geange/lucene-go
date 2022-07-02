package analysis

// CachingTokenFilter This class can be used if the token attributes of a TokenStream are intended to be
// consumed more than once. It caches all token attribute states locally in a List when the first call to
// incrementToken() is called. Subsequent calls will used the cache.
// Important: Like any proper TokenFilter, reset() propagates to the input, although only before incrementToken()
// is called the first time. Prior to Lucene 5, it was never propagated.
type CachingTokenFilter struct {
}
