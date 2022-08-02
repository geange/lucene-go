package store

import "io"

// Lock An interprocess mutex lock.
// Typical use might look like:
//     try (final Lock lock = directory.obtainLock("my.lock")) {
//       // ... code to execute while locked ...
//     }
//
// See Also: Directory.obtainLock(String)
type Lock interface {
	// Closer Releases exclusive access.
	// Note that exceptions thrown from close may require human intervention, as it may mean the lock was no longer valid, or that fs permissions prevent removal of the lock file, or other reasons.
	// Closes this stream and releases any system resources associated with it. If the stream is already closed then invoking this method has no effect.
	// As noted in AutoCloseable.close(), cases where the close may fail require careful attention. It is strongly advised to relinquish the underlying resources and to internally mark the Closeable as closed, prior to throwing the IOException.
	// Throws: LockReleaseFailedException – optional specific exception) if the lock could not be properly released.
	io.Closer

	// EnsureValid Best effort check that this lock is still valid. Locks could become invalidated externally for a number of reasons, for example if a user deletes the lock file manually or when a network filesystem is in use.
	// Throws: IOException – if the lock is no longer valid.
	EnsureValid() error
}
