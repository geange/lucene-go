package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesRef(t *testing.T) {
	bs := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	bytesRef := NewBytesRef(bs)
	err := bytesRef.Set(1, 2)
	assert.Nil(t, err, nil)
	assert.Equal(t, bytesRef.Bytes(), []byte{1, 2})

	err = bytesRef.Set(8, 2)
	assert.Nil(t, err, nil)
	assert.Equal(t, bytesRef.Bytes(), []byte{8, 9})

	err = bytesRef.Set(8, 3)
	assert.NotNil(t, err)
}
