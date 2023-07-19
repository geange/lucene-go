package store

import "io"

// Lock An interprocess mutex lock.
// Typical use might look like:
//
//	try (final Lock lock = directory.obtainLock("my.lock")) {
//	  // ... code to execute while locked ...
//	}
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

// LockFactory Base class for Locking implementation. Directory uses instances of this class to implement locking.
//
// Lucene uses NativeFSLockFactory by default for FSDirectory-based index directories.
//
// Special care needs to be taken if you change the locking implementation: First be certain that no writer is
// in fact writing to the index otherwise you can easily corrupt your index. Be sure to do the LockFactory
// change on all Lucene instances and clean up all leftover lock files before starting the new configuration
// for the first time. Different implementations can not work together!
//
// If you suspect that some LockFactory implementation is not working properly in your environment, you can
// easily test it by using VerifyingLockFactory, LockVerifyServer and LockStressTest.
//
// See Also: 	LockVerifyServer,
//
//	LockStressTest,
//	VerifyingLockFactory
type LockFactory interface {

	// ObtainLock Return a new obtained Lock instance identified by lockName.
	// Params: lockName – name of the lock to be created.
	// Throws: 	LockObtainFailedException – (optional specific exception) if the lock could not be obtained
	//			because it is currently held elsewhere.
	//			IOException – if any i/o error occurs attempting to gain the lock
	ObtainLock(dir Directory, lockName string) (Lock, error)
}
