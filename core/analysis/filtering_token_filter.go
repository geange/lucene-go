package analysis

// FilteringTokenFilter Abstract base class for TokenFilters that may remove tokens. You have to implement
// accept and return a boolean if the current token should be preserved. incrementToken uses this method to
// decide if a token should be passed to the caller.
type FilteringTokenFilter struct {
}
