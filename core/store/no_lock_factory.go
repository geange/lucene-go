package store

// NoLockFactory Use this LockFactory to disable locking entirely. This is a singleton, you have to use INSTANCE.
// See Also: LockFactory
type NoLockFactory interface {
	LockFactory
}
