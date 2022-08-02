package store

var _ LockFactory = &SingleInstanceLockFactory{}

// SingleInstanceLockFactory Implements LockFactory for a single in-process instance, meaning all
// locking will take place through this one instance. Only use this LockFactory when you are certain
// all IndexWriters for a given index are running against a single shared in-process Directory instance.
// This is currently the default locking for RAMDirectory.
// See Also: LockFactory
type SingleInstanceLockFactory struct {
}
