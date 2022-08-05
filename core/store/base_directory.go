package store

// BaseDirectory Base implementation for a concrete Directory that uses a LockFactory for locking.
type BaseDirectory interface {
	Directory
}

type BaseDirectoryImp struct {
	isOpen bool

	// Holds the LockFactory instance (implements locking for this Directory instance).
	lockFactory LockFactory
}

func (b *BaseDirectoryImp) ObtainLock(name string) (Lock, error) {
	//return b.lockFactory.ObtainLock(, name)
	panic("")
}
