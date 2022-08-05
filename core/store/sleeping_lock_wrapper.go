package store

//var _ FilterDirectory = &SleepingLockWrapper{}

// SleepingLockWrapper Directory that wraps another, and that sleeps and retries if obtaining the lock fails.
// This is not a good idea.
type SleepingLockWrapper struct {
}
