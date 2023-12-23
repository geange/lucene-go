package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNoLockFactory(t *testing.T) {
	lockFactory := NewNoLockFactory()
	lock, err := lockFactory.ObtainLock(nil, "")
	assert.Nil(t, err)

	err = lock.EnsureValid()
	assert.Nil(t, err)

	err = lock.Close()
	assert.Nil(t, err)
}
