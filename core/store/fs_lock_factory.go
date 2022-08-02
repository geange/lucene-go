package store

// FSLockFactory Base class for file system based locking implementation. This class is explicitly
// checking that the passed Directory is an FSDirectory.
type FSLockFactory interface {
	LockFactory
}
