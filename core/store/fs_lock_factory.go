package store

import "errors"

// FSLockFactory Base class for file system based locking implementation. This class is explicitly
// checking that the passed Directory is an FSDirectory.
type FSLockFactory interface {
	LockFactory

	// ObtainFSLock Implement this method to obtain a lock for a FSDirectory instance.
	// Throws: IOException – if the lock could not be obtained.
	ObtainFSLock(dir FSDirectory, lockName string) (Lock, error)
}

type FSLockFactoryNeed interface {
	ObtainFSLock(dir FSDirectory, lockName string) (Lock, error)
}

var _ FSLockFactory = &FSLockFactoryImp{}

type FSLockFactoryImp struct {
	FSLockFactoryNeed
}

func NewFSLockFactoryImp(need FSLockFactoryNeed) *FSLockFactoryImp {
	return &FSLockFactoryImp{FSLockFactoryNeed: need}
}

func (f *FSLockFactoryImp) ObtainLock(dir Directory, lockName string) (Lock, error) {
	fsDir, ok := dir.(FSDirectory)
	if !ok {
		return nil, errors.New("can only be used with FSDirectory subclasses")
	}
	return f.ObtainFSLock(fsDir, lockName)
}
