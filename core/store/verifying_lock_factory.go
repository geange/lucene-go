package store

var _ LockFactory = &VerifyingLockFactory{}

// VerifyingLockFactory A LockFactory that wraps another LockFactory and verifies that each lock
// obtain/release is "correct" (never results in two processes holding the lock at the same time).
// It does this by contacting an external server (LockVerifyServer) to assert that at most one process
// holds the lock at a time. To use this, you should also run LockVerifyServer on the host and port
// matching what you pass to the constructor.
//
// See Also: 	LockVerifyServer,
//				LockStressTest
type VerifyingLockFactory struct {
}
