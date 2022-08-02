package store

var _ FSLockFactory = &SimpleFSLockFactory{}

// SimpleFSLockFactory Implements LockFactory using Files.createFile.
// The main downside with using this API for locking is that the Lucene write lock may not be released when the JVM exits abnormally.
// When this happens, an LockObtainFailedException is hit when trying to create a writer, in which case you may need to explicitly clear the lock file first by manually removing the file. But, first be certain that no writer is in fact writing to the index otherwise you can easily corrupt your index.
// Special care needs to be taken if you change the locking implementation: First be certain that no writer is in fact writing to the index otherwise you can easily corrupt your index. Be sure to do the LockFactory change all Lucene instances and clean up all leftover lock files before starting the new configuration for the first time. Different implementations can not work together!
// If you suspect that this or any other LockFactory is not working properly in your environment, you can easily test it by using VerifyingLockFactory, LockVerifyServer and LockStressTest.
// This is a singleton, you have to use INSTANCE.
// See Also: LockFactory
type SimpleFSLockFactory struct {
}
