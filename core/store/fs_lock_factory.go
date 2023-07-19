package store

import "errors"

// FSLockFactory Base class for file system based locking implementation. This class is explicitly
// checking that the passed Directory is an FSDirectory.
type FSLockFactory interface {
	LockFactory

	// FSLockFactoryInner Implement this method to obtain a lock for a FSDirectory instance.
	// Throws: IOException – if the lock could not be obtained.
	FSLockFactoryInner
}

type FSLockFactoryInner interface {
	ObtainFSLock(dir FSDirectory, lockName string) (Lock, error)
}

type FSLockFactoryBase struct {
	locker FSLockFactoryInner
}

func NewFSLockFactoryImp(inner FSLockFactoryInner) *FSLockFactoryBase {
	return &FSLockFactoryBase{locker: inner}
}

func (f *FSLockFactoryBase) ObtainLock(dir Directory, lockName string) (Lock, error) {
	fsDir, ok := dir.(FSDirectory)
	if !ok {
		return nil, errors.New("can only be used with FSDirectory subclasses")
	}
	return f.locker.ObtainFSLock(fsDir, lockName)
}
