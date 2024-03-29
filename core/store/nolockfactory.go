package store

var _ LockFactory = &NoLockFactory{}

var NoLockFactoryInstance = NewNoLockFactory()

// NoLockFactory Use this LockFactory to disable locking entirely. This is a singleton, you have to use INSTANCE.
// See Also: LockFactory
type NoLockFactory struct {
	singletonLock *NoLock
}

func NewNoLockFactory() *NoLockFactory {
	return &NoLockFactory{singletonLock: &NoLock{}}
}

func (n *NoLockFactory) ObtainLock(dir Directory, lockName string) (Lock, error) {
	return n.singletonLock, nil
}

var _ Lock = &NoLock{}

type NoLock struct {
}

func (n *NoLock) Close() error {
	return nil
}

func (n *NoLock) EnsureValid() error {
	return nil
}
