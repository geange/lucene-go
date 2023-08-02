package store

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomAccessIndexInput(t *testing.T) {
	output := NewBytesInput([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0})
	defer output.Close()

	access := &randomAccessIndexInput{in: output}

	rUint8, err := access.ReadU8(0)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, rUint8)

	bs := make([]byte, 4)
	n, err := access.ReadAt(bs, 7)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, n)

	rUint16, err := access.ReadU16(0)
	assert.Nil(t, err)
	assert.EqualValues(t, binary.BigEndian.Uint16([]byte{1, 2}), rUint16)

	rUint32, err := access.ReadU32(0)
	assert.Nil(t, err)
	assert.EqualValues(t, binary.BigEndian.Uint32([]byte{1, 2, 3, 4}), rUint32)

	rUint64, err := access.ReadU64(0)
	assert.Nil(t, err)
	assert.EqualValues(t, binary.BigEndian.Uint64([]byte{1, 2, 3, 4, 5, 6, 7, 8}), rUint64)
}

func TestBytesRandomAccessInput(t *testing.T) {
	access := NewBytesRandomAccessInput([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, binary.BigEndian)

	rUint8, err := access.ReadU8(0)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, rUint8)

	bs := make([]byte, 4)
	n, err := access.ReadAt(bs, 7)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, n)

	n, err = access.ReadAt(bs, 0)
	assert.Nil(t, err)
	assert.EqualValues(t, 4, n)

	_, err = access.ReadAt(bs, 100)
	assert.NotNil(t, err)

	rUint16, err := access.ReadU16(0)
	assert.Nil(t, err)
	assert.EqualValues(t, binary.BigEndian.Uint16([]byte{1, 2}), rUint16)

	rUint32, err := access.ReadU32(0)
	assert.Nil(t, err)
	assert.EqualValues(t, binary.BigEndian.Uint32([]byte{1, 2, 3, 4}), rUint32)

	rUint64, err := access.ReadU64(0)
	assert.Nil(t, err)
	assert.EqualValues(t, binary.BigEndian.Uint64([]byte{1, 2, 3, 4, 5, 6, 7, 8}), rUint64)

}
