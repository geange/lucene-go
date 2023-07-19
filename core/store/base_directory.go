package store

import "errors"

// BaseDirectory Base implementation for a concrete Directory that uses a LockFactory for locking.
type BaseDirectory interface {
	Directory
}

type BaseDirectoryBase struct {
	dir    Directory
	isOpen bool

	// Holds the LockFactory instance (implements locking for this Directory instance).
	lockFactory LockFactory
}

func NewBaseDirectoryBase(dir Directory, lockFactory LockFactory) *BaseDirectoryBase {
	return &BaseDirectoryBase{
		dir:         dir,
		isOpen:      true,
		lockFactory: lockFactory,
	}
}

func (b *BaseDirectoryBase) ObtainLock(name string) (Lock, error) {
	return b.lockFactory.ObtainLock(b.dir, name)
}

func (b *BaseDirectoryBase) EnsureOpen() error {
	if !b.isOpen {
		return errors.New("this Directory is closed")
	}
	return nil
}
