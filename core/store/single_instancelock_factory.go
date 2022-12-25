package store

import (
	"fmt"
	"sync"
)

var _ LockFactory = &SingleInstanceLockFactory{}

// SingleInstanceLockFactory Implements LockFactory for a single in-process instance, meaning all
// locking will take place through this one instance. Only use this LockFactory when you are certain
// all IndexWriters for a given index are running against a single shared in-process Directory instance.
// This is currently the default locking for RAMDirectory.
// See Also: LockFactory
type SingleInstanceLockFactory struct {
	sync.RWMutex
	locks map[string]struct{}
}

func (s *SingleInstanceLockFactory) ObtainLock(dir Directory, lockName string) (Lock, error) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.locks[lockName]; ok {
		return nil, fmt.Errorf(
			"lock instance already obtained: (dir=%s, lockName=%s)",
			dir, lockName)
	}
	return NewSingleInstanceLock(s, lockName), nil
}

var _ Lock = &SingleInstanceLock{}

type SingleInstanceLock struct {
	*SingleInstanceLockFactory

	lockName string
	closed   bool
}

func NewSingleInstanceLock(factory *SingleInstanceLockFactory, lockName string) *SingleInstanceLock {
	return &SingleInstanceLock{SingleInstanceLockFactory: factory, lockName: lockName}
}

func (s *SingleInstanceLock) Close() error {
	if s.closed {
		return nil
	}

	s.Lock()
	defer s.Unlock()

	if _, ok := s.locks[s.lockName]; !ok {
		return fmt.Errorf("lock was already released: %s", s.lockName)
	}

	delete(s.locks, s.lockName)
	s.closed = true
	return nil
}

func (s *SingleInstanceLock) EnsureValid() error {
	if s.closed {
		return fmt.Errorf("lock instance already released: %s", s.lockName)
	}

	s.Lock()
	defer s.Unlock()

	if _, ok := s.locks[s.lockName]; !ok {
		return fmt.Errorf("lock instance was invalidated from map: %s", s.lockName)
	}

	return nil
}
