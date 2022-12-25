package store

// RateLimiter Abstract base class to rate limit IO. Typically implementations are shared across multiple
// IndexInputs or IndexOutputs (for example those involved all merging). Those IndexInputs and IndexOutputs
// would call pause whenever the have read or written more than getMinPauseCheckBytes bytes.
type RateLimiter interface {
}
