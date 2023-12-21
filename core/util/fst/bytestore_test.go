package fst

import (
	"bytes"
	"context"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestByteStoreCopyTo(t *testing.T) {

	bstore := NewByteStore(10)

	_, err := bstore.Write([]byte("1111111112"))
	assert.Nil(t, err)

	_, err = bstore.Write([]byte("Aaaaaaaaab"))
	assert.Nil(t, err)

	buf := new(bytes.Buffer)
	err = bstore.CopyTo(context.Background(), 0, 10, buf)
	assert.Nil(t, err)
	assert.Equal(t, "1111111112", buf.String())

	buf.Reset()
	err = bstore.CopyTo(context.Background(), 10, 10, buf)
	assert.Nil(t, err)
	assert.Equal(t, "Aaaaaaaaab", buf.String())
}

func TestByteStoreWriteInt32(t *testing.T) {

	bstore := NewByteStore(10)

	_, err := bstore.Write(make([]byte, 20))
	assert.Nil(t, err)

	err = bstore.WriteInt32(0, 100)
	assert.Nil(t, err)

	buf := new(bytes.Buffer)
	err = bstore.CopyTo(context.Background(), 0, 4, buf)
	assert.Nil(t, err)
	n := int(binary.BigEndian.Uint32(buf.Bytes()))
	assert.Equal(t, 100, n)
}

func TestByteStoreSkipBytes(t *testing.T) {

	bstore := NewByteStore(10)

	_, err := bstore.Write(make([]byte, 20))
	assert.Nil(t, err)

	size := int64(1 << 12)

	err = bstore.SkipBytes(size)
	assert.Nil(t, err)

	assert.Equal(t, size+20, bstore.GetPosition())
}

func TestByteStoreTruncate(t *testing.T) {

	{
		bstore := NewByteStore(10)

		_, err := bstore.Write(make([]byte, 20))
		assert.Nil(t, err)

		err = bstore.SkipBytes(1 << 11)
		assert.Nil(t, err)

		assert.Equal(t, int64((1<<10)*2+20), bstore.GetPosition())

		err = bstore.Truncate(20)
		assert.Nil(t, err)

		assert.Equal(t, int64(20), bstore.GetPosition())
	}

	{
		// downTo == 0
		bstore := NewByteStore(10)

		_, err := bstore.Write(make([]byte, 20))
		assert.Nil(t, err)

		err = bstore.SkipBytes(1 << 11)
		assert.Nil(t, err)

		assert.Equal(t, int64((1<<10)*2+20), bstore.GetPosition())

		err = bstore.Truncate(20)
		assert.Nil(t, err)

		assert.Equal(t, int64(20), bstore.GetPosition())
	}

	{
		bstore := NewByteStore(10)

		_, err := bstore.Write(make([]byte, 20))
		assert.Nil(t, err)

		err = bstore.SkipBytes(1 << 11)
		assert.Nil(t, err)

		assert.Equal(t, int64((1<<10)*2+20), bstore.GetPosition())

		err = bstore.Truncate(int64((1<<10)*2 + 20))
		assert.Nil(t, err)

		assert.Equal(t, int64((1<<10)*2+20), bstore.GetPosition())
	}
}

func TestByteStoreWrite(t *testing.T) {
	bstore := NewByteStore(10)

	_, err := bstore.Write(make([]byte, 1048))
	assert.Nil(t, err)

	assert.Equal(t, int64(1048), bstore.GetPosition())
}

func TestByteStoreWriteByteAt(t *testing.T) {
	bstore := NewByteStore(10)

	_, err := bstore.Write(make([]byte, 1048))
	assert.Nil(t, err)

	err = bstore.WriteByteAt(0, 'a')
	assert.Nil(t, err)

	buf := new(bytes.Buffer)
	err = bstore.CopyTo(context.Background(), 0, 1, buf)
	assert.Nil(t, err)

	assert.Equal(t, []byte{'a'}, buf.Bytes())
}

func TestByteStoreMoveBytes(t *testing.T) {
	{
		bstore := NewByteStore(10)

		data := [1048]byte{1, 2, 3, 4}

		_, err := bstore.Write(data[:])
		assert.Nil(t, err)

		err = bstore.MoveBytes(context.Background(), 0, 1023, 4)
		assert.Nil(t, err)

		buf := new(bytes.Buffer)
		err = bstore.CopyTo(context.Background(), 1023, 4, buf)
		assert.Nil(t, err)

		assert.Equal(t, []byte{1, 2, 3, 4}, buf.Bytes())
	}

	{
		bstore := NewByteStore(10)

		data := [2048]byte{1, 2, 3, 4}

		_, err := bstore.Write(data[:])
		assert.Nil(t, err)

		err = bstore.MoveBytes(context.Background(), 0, 1024, 1024)
		assert.Nil(t, err)

		buf := new(bytes.Buffer)
		err = bstore.CopyTo(context.Background(), 1024, 4, buf)
		assert.Nil(t, err)

		assert.Equal(t, []byte{1, 2, 3, 4}, buf.Bytes())
	}

	{
		bstore := NewByteStore(10)

		data := [1024]byte{1, 2, 3, 4}
		data[1020] = 5
		data[1021] = 6
		data[1022] = 7
		data[1023] = 8

		err := bstore.SkipBytes(4096)
		assert.Nil(t, err)

		err = bstore.WriteBytesAt(context.Background(), 0, data[:])
		assert.Nil(t, err)

		err = bstore.MoveBytes(context.Background(), 0, 1023, 1048)
		assert.Nil(t, err)

		buf := new(bytes.Buffer)
		err = bstore.CopyTo(context.Background(), 1023+1024-4, 4, buf)
		assert.Nil(t, err)
		assert.Equal(t, []byte{5, 6, 7, 8}, buf.Bytes())

		buf.Reset()
		err = bstore.CopyTo(context.Background(), 1023, 4, buf)
		assert.Nil(t, err)
		assert.Equal(t, []byte{1, 2, 3, 4}, buf.Bytes())
	}
}
