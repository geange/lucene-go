package store

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"hash/crc32"
	"testing"
)

func TestNewHash(t *testing.T) {
	items := [][]byte{
		make([]byte, 100),
		[]byte("12345"),
		{1, 2, 4, 5},
		{0, 0, 9},
	}

	stdHash := crc32.NewIEEE()
	customHash := NewHash()

	for _, item := range items {
		stdHash.Write(item)
		customHash.Write(item)
		assert.EqualValues(t, customHash.Sum(), stdHash.Sum32())
	}
}

func TestNewHashClone(t *testing.T) {
	hash := NewHash()

	hash.Write(make([]byte, 100))
	code := hash.Sum()

	cloneHash := hash.Clone()

	data := [100]byte{1, 2, 3}
	hash.Write(data[:])

	assert.EqualValues(t, code, cloneHash.Sum())

	cloneHash.Write(data[:])
	assert.EqualValues(t, hash.Sum(), cloneHash.Sum())
}

func TestBufferedChecksumIndexOutput(t *testing.T) {
	buf := new(bytes.Buffer)

	items := [][]byte{
		make([]byte, 100),
		{1},
		{2},
		{3},
		{4},
	}

	hash := NewHash()

	for _, item := range items {
		hash.Write(item)
		buf.Write(item)
	}

	input := NewBufferDataInput(buf)
	indexInput := NewBufferedChecksumIndexInput(input)

	bs := make([]byte, 200)
	n, err := indexInput.Read(bs)
	assert.Nil(t, err)
	assert.EqualValues(t, 100+4, n)
	assert.EqualValues(t, hash.Sum(), indexInput.GetChecksum())
}

func TestBufferedChecksumIndexOutputClone(t *testing.T) {
	buf := new(bytes.Buffer)

	items := [][]byte{
		make([]byte, 100),
		{1},
		{2},
		{3},
		{4},
	}

	for _, item := range items {
		buf.Write(item)
	}

	input := NewBufferDataInput(buf)
	indexInput := NewBufferedChecksumIndexInput(input)

	cloneIndexInput := indexInput.Clone().(*BufferedChecksumIndexInput)

	bs := make([]byte, 200)
	n, err := indexInput.Read(bs)
	assert.Nil(t, err)
	assert.EqualValues(t, 100+4, n)

	n, err = cloneIndexInput.Read(bs)
	assert.Nil(t, err)
	assert.EqualValues(t, 100+4, n)

	assert.EqualValues(t, cloneIndexInput.GetChecksum(), indexInput.GetChecksum())
}
