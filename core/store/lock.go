package store

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Lock An interprocess mutex lock.
// See Also: Directory.obtainLock(String)
type Lock interface {
	// Closer Releases exclusive access.
	// Note that exceptions thrown from close may require human intervention, as it may mean the lock was no longer valid, or that fs permissions prevent removal of the lock file, or other reasons.
	// Closes this stream and releases any system resources associated with it. If the stream is already closed then invoking this method has no effect.
	// As noted in AutoCloseable.close(), cases where the close may fail require careful attention. It is strongly advised to relinquish the underlying resources and to internally mark the Closeable as closed, prior to throwing the IOException.
	// Throws: LockReleaseFailedException – optional specific exception) if the lock could not be properly released.
	io.Closer

	// EnsureValid Best effort check that this lock is still valid. Locks could become invalidated externally for a number of reasons, for example if a user deletes the lock file manually or when a network filesystem is in use.
	// Throws: IOException – if the lock is no longer valid.
	EnsureValid() error
}

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
//
//	LockStressTest,
//	VerifyingLockFactory
type LockFactory interface {

	// ObtainLock Return a new obtained Lock instance identified by lockName.
	// Params: lockName – name of the lock to be created.
	// Throws: 	LockObtainFailedException – (optional specific exception) if the lock could not be obtained
	//			because it is currently held elsewhere.
	//			IOException – if any i/o error occurs attempting to gain the lock
	ObtainLock(dir Directory, lockName string) (Lock, error)
}

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

func NewFSLockFactoryBase(inner FSLockFactoryInner) *FSLockFactoryBase {
	return &FSLockFactoryBase{locker: inner}
}

func (f *FSLockFactoryBase) ObtainLock(dir Directory, lockName string) (Lock, error) {
	fsDir, ok := dir.(FSDirectory)
	if !ok {
		return nil, errors.New("can only be used with FSDirectory subclasses")
	}
	return f.locker.ObtainFSLock(fsDir, lockName)
}

var _ FSLockFactory = &SimpleFSLockFactory{}

// SimpleFSLockFactory Implements LockFactory using Files.createFile.
// The main downside with using this API for locking is that the Lucene write lock may not be released when the JVM exits abnormally.
// When this happens, an LockObtainFailedException is hit when trying to create a writer, in which case you may need to explicitly clear the lock file first by manually removing the file. But, first be certain that no writer is in fact writing to the index otherwise you can easily corrupt your index.
// Special care needs to be taken if you change the locking implementation: First be certain that no writer is in fact writing to the index otherwise you can easily corrupt your index. Be sure to do the LockFactory change all Lucene instances and clean up all leftover lock files before starting the new configuration for the first time. Different implementations can not work together!
// If you suspect that this or any other LockFactory is not working properly in your environment, you can easily test it by using VerifyingLockFactory, LockVerifyServer and LockStressTest.
// This is a singleton, you have to use INSTANCE.
// See Also: LockFactory
type SimpleFSLockFactory struct {
	*FSLockFactoryBase
}

func NewSimpleFSLockFactory() *SimpleFSLockFactory {
	factory := &SimpleFSLockFactory{}
	factory.FSLockFactoryBase = NewFSLockFactoryBase(factory)
	return factory
}

func (s *SimpleFSLockFactory) ObtainFSLock(dir FSDirectory, lockName string) (Lock, error) {
	lockDir, err := dir.GetDirectory()
	if err != nil {
		return nil, err
	}

	// Ensure that lockDir exists and is a directory.
	// note: this will fail if lockDir is a symlink
	// Files.createDirectories(lockDir)

	lockFile := filepath.Join(lockDir, lockName)

	// create the file: this will fail if it already exists
	if _, err = os.Stat(lockFile); err == nil {
		return nil, fmt.Errorf("lock held elsewhere: %s", lockFile)
	}

	if os.IsNotExist(err) {
		// 文件不存在
		if _, err := os.Create(lockFile); err != nil {
			return nil, err
		}
		info, err := os.Stat(lockFile)
		if err != nil {
			return nil, err
		}
		_, ctime, _ := FileTime(info)
		return NewSimpleFSLock(lockFile, ctime), nil
	}

	return nil, err
}

var _ Lock = &SimpleFSLock{}

type SimpleFSLock struct {
	path         string
	creationTime time.Time
	closed       bool
}

func NewSimpleFSLock(path string, creationTime time.Time) *SimpleFSLock {
	return &SimpleFSLock{path: path, creationTime: creationTime}
}

func (s *SimpleFSLock) Close() error {
	return os.Remove(s.path)
}

func (s *SimpleFSLock) EnsureValid() error {
	if s.closed {
		return fmt.Errorf("lock instance already released: %s", s.path)
	}
	info, err := os.Stat(s.path)
	if err != nil {
		return err
	}
	_, ctime, _ := FileTime(info)
	if !s.creationTime.Equal(ctime) {
		return fmt.Errorf(
			"underlying file changed by an external force at %s, (lock=%s)",
			ctime.String(), s.path)
	}
	return nil
}

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

func NewSingleInstanceLockFactory() *SingleInstanceLockFactory {
	return &SingleInstanceLockFactory{
		RWMutex: sync.RWMutex{},
		locks:   make(map[string]struct{}),
	}
}

func (s *SingleInstanceLockFactory) ObtainLock(_ Directory, lockName string) (Lock, error) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.locks[lockName]; ok {
		return nil, fmt.Errorf("lock instance already obtained: (lockName=%s)", lockName)
	}
	s.locks[lockName] = struct{}{}
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

/*

const (
	MSG_LOCK_RELEASED = 0
	MSG_LOCK_ACQUIRED = 1
)

var _ LockFactory = &VerifyingLockFactory{}

// VerifyingLockFactory A LockFactory that wraps another LockFactory and verifies that each lock
// obtain/release is "correct" (never results in two processes holding the lock at the same time).
// It does this by contacting an external server (LockVerifyServer) to assert that at most one process
// holds the lock at a time. To use this, you should also run LockVerifyServer on the host and port
// matching what you pass to the constructor.
//
// See Also: 	LockVerifyServer,
//
//	LockStressTest
type VerifyingLockFactory struct {
	lf  LockFactory
	in  io.ByteReader
	out io.ByteWriter
}

// NewVerifyingLockFactory
// lf: the LockFactory that we are testing
// in: the socket's input to LockVerifyServer
// out: the socket's output to LockVerifyServer
func NewVerifyingLockFactory(lf LockFactory, in io.ByteReader, out io.ByteWriter) *VerifyingLockFactory {
	return &VerifyingLockFactory{lf: lf, in: in, out: out}
}

func (v *VerifyingLockFactory) ObtainLock(dir Directory, lockName string) (Lock, error) {
	lock, err := v.lf.ObtainLock(dir, lockName)
	if err != nil {
		return nil, err
	}
	return NewCheckedLock(v, lock), nil
}

var _ Lock = &CheckedLock{}

type CheckedLock struct {
	*VerifyingLockFactory

	lock Lock

	buff []byte
}

func NewCheckedLock(factory *VerifyingLockFactory, lock Lock) *CheckedLock {
	return &CheckedLock{
		VerifyingLockFactory: factory,
		lock:                 lock,
		buff:                 make([]byte, 1),
	}
}

func (c *CheckedLock) Close() error {
	if err := c.lock.EnsureValid(); err != nil {
		return err
	}
	return c.verify(MSG_LOCK_RELEASED)
}

func (c *CheckedLock) EnsureValid() error {
	return c.lock.EnsureValid()
}

func (c *CheckedLock) verify(message byte) error {
	err := c.out.WriteByte(message)
	if err != nil {
		return err
	}

	ret, err := c.in.ReadByte()
	if err != nil {
		return err
	}
	if ret < 0 {
		return errors.New("lock server died because of locking error")
	}
	if ret != message {
		return errors.New("protocol violation")
	}
	return nil
}

*/
