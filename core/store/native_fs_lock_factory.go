package store

var _ FSLockFactory = &NativeFSLockFactory{}

// NativeFSLockFactory Implements LockFactory using native OS file locks. Note that because this LockFactory relies on java.nio.* APIs for locking, any problems with those APIs will cause locking to fail. Specifically, on certain NFS environments the java.nio.* locks will fail (the lock can incorrectly be double acquired) whereas SimpleFSLockFactory worked perfectly in those same environments. For NFS based access to an index, it's recommended that you try SimpleFSLockFactory first and work around the one limitation that a lock file could be left when the JVM exits abnormally.
// The primary benefit of NativeFSLockFactory is that locks (not the lock file itself) will be properly removed (by the OS) if the JVM has an abnormal exit.
// Note that, unlike SimpleFSLockFactory, the existence of leftover lock files in the filesystem is fine because the OS will free the locks held against these files even though the files still remain. Lucene will never actively remove the lock files, so although you see them, the index may not be locked.
// Special care needs to be taken if you change the locking implementation: First be certain that no writer is in fact writing to the index otherwise you can easily corrupt your index. Be sure to do the LockFactory change on all Lucene instances and clean up all leftover lock files before starting the new configuration for the first time. Different implementations can not work together!
// If you suspect that this or any other LockFactory is not working properly in your environment, you can easily test it by using VerifyingLockFactory, LockVerifyServer and LockStressTest.
// This is a singleton, you have to use INSTANCE.
// See Also: LockFactory
type NativeFSLockFactory struct {
}
