package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleFSLock(t *testing.T) {
	dir, err := NewNIOFSDirectory(os.TempDir())
	if assert.Nil(t, err) {
		defer dir.Close()
	}

	lockName := "xx"

	lock, err := dir.ObtainLock(lockName)
	assert.Nil(t, err)

	err = lock.EnsureValid()
	assert.Nil(t, err)

	err = lock.Close()
	assert.Nil(t, err)

	_, err = os.Open(filepath.Join(os.TempDir(), lockName))
	assert.NotNil(t, err)
}

func TestSingleInstanceLockFactory(t *testing.T) {
	{
		factory := NewSingleInstanceLockFactory()
		lock, err := factory.ObtainLock(nil, "xx")
		assert.Nil(t, err)

		err = lock.EnsureValid()
		assert.Nil(t, err)

		err = lock.Close()
		assert.Nil(t, err)
	}

	{
		factory := NewSingleInstanceLockFactory()
		lock, err := factory.ObtainLock(nil, "xx")
		assert.Nil(t, err)

		_, err = factory.ObtainLock(nil, "xx")
		assert.NotNil(t, err)

		err = lock.EnsureValid()
		assert.Nil(t, err)

		err = lock.Close()
		assert.Nil(t, err)
	}

	{
		factory := NewSingleInstanceLockFactory()
		lock, err := factory.ObtainLock(nil, "xx")
		assert.Nil(t, err)

		err = lock.EnsureValid()
		assert.Nil(t, err)

		err = lock.Close()
		assert.Nil(t, err)

		newLock, err := factory.ObtainLock(nil, "xx")
		assert.Nil(t, err)

		err = newLock.Close()
		assert.Nil(t, err)
	}
}
