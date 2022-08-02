package store

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
//				LockStressTest,
//				VerifyingLockFactory
type LockFactory interface {
}
