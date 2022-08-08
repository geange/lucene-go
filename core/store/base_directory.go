package store

// BaseDirectory Base implementation for a concrete Directory that uses a LockFactory for locking.
type BaseDirectory interface {
	Directory
}

type BaseDirectoryImp struct {
	dir Directory

	isOpen bool

	// Holds the LockFactory instance (implements locking for this Directory instance).
	lockFactory LockFactory
}

func NewBaseDirectoryImp(dir Directory, lockFactory LockFactory) *BaseDirectoryImp {
	return &BaseDirectoryImp{
		dir:         dir,
		isOpen:      true,
		lockFactory: lockFactory,
	}
}

func (b *BaseDirectoryImp) ObtainLock(name string) (Lock, error) {
	return b.lockFactory.ObtainLock(b.dir, name)
}
