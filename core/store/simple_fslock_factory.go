package store

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var _ FSLockFactory = &SimpleFSLockFactory{}

// SimpleFSLockFactory Implements LockFactory using Files.createFile.
// The main downside with using this API for locking is that the Lucene write lock may not be released when the JVM exits abnormally.
// When this happens, an LockObtainFailedException is hit when trying to create a writer, in which case you may need to explicitly clear the lock file first by manually removing the file. But, first be certain that no writer is in fact writing to the index otherwise you can easily corrupt your index.
// Special care needs to be taken if you change the locking implementation: First be certain that no writer is in fact writing to the index otherwise you can easily corrupt your index. Be sure to do the LockFactory change all Lucene instances and clean up all leftover lock files before starting the new configuration for the first time. Different implementations can not work together!
// If you suspect that this or any other LockFactory is not working properly in your environment, you can easily test it by using VerifyingLockFactory, LockVerifyServer and LockStressTest.
// This is a singleton, you have to use INSTANCE.
// See Also: LockFactory
type SimpleFSLockFactory struct {
	*FSLockFactoryImp
}

func NewSimpleFSLockFactory() *SimpleFSLockFactory {
	factory := &SimpleFSLockFactory{}
	factory.FSLockFactoryImp = NewFSLockFactoryImp(factory)
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
	_, err = os.Stat(lockFile)
	if err == nil {
		return nil, fmt.Errorf("lock held elsewhere: %s", lockFile)
	}

	if os.IsNotExist(err) {
		// 文件不存在
		_, err := os.Create(lockFile)
		if err != nil {
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
	//TODO implement me
	panic("implement me")
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
