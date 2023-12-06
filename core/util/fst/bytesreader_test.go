package fst

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"

	"github.com/geange/lucene-go/core/store"
	"github.com/stretchr/testify/assert"
)

func TestReverseBytesReader(t *testing.T) {
	bs := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	reader := newReverseBytesReader(bs)

	reversed := reader.Reversed()
	assert.True(t, reversed)

	b1, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, b1, byte(9))

	position1 := reader.GetPosition()
	assert.Equal(t, int64(8), position1)

	b2, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, b2, byte(8))

	position2 := reader.GetPosition()
	assert.Equal(t, int64(7), position2)

	err = reader.SkipBytes(context.TODO(), 2)
	assert.Nil(t, err)

	position3 := reader.GetPosition()
	assert.Equal(t, int64(5), position3)

	b3, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, b3, byte(5))

	err = reader.SetPosition(int64(len(bs) - 1))
	assert.Nil(t, err)

	p := make([]byte, 4)
	n, err := reader.Read(p)
	assert.Nil(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte{9, 8, 7, 6}, p)

}

func TestReverseRandomAccessReader(t *testing.T) {
	bs := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	input := store.NewBytesRandomAccessInput(bs, binary.BigEndian)

	reader := newReverseRandomAccessReader(input)

	err := reader.SetPosition(int64(len(bs) - 1))
	assert.Nil(t, err)

	reversed := reader.Reversed()
	assert.True(t, reversed)

	b1, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, b1, byte(9))

	position1 := reader.GetPosition()
	assert.Equal(t, int64(8), position1)

	b2, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, b2, byte(8))

	position2 := reader.GetPosition()
	assert.Equal(t, int64(7), position2)

	err = reader.SkipBytes(context.TODO(), 2)
	assert.Nil(t, err)

	position3 := reader.GetPosition()
	assert.Equal(t, int64(5), position3)

	b3, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, b3, byte(5))

	err = reader.SetPosition(int64(len(bs) - 1))
	assert.Nil(t, err)

	p := make([]byte, 4)
	n, err := reader.Read(p)
	assert.Nil(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte{9, 8, 7, 6}, p)

}

func TestBuilderBytesReader(t *testing.T) {
	bs := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	input, err := NewBytesStoreByDataInput(bytes.NewReader(bs), 10, 1<<10)
	assert.Nil(t, err)

	reader, err := newBuilderBytesReader(input)
	assert.Nil(t, err)

	reversed := reader.Reversed()
	assert.True(t, reversed)

	err = reader.SetPosition(int64(len(bs) - 1))
	assert.Nil(t, err)

	b1, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, b1, byte(9))

	position1 := reader.GetPosition()
	assert.Equal(t, int64(8), position1)

	b2, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, b2, byte(8))

	position2 := reader.GetPosition()
	assert.Equal(t, int64(7), position2)

	err = reader.SkipBytes(context.TODO(), 2)
	assert.Nil(t, err)

	position3 := reader.GetPosition()
	assert.Equal(t, int64(5), position3)

	b3, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, b3, byte(5))

	err = reader.SetPosition(int64(len(bs) - 1))
	assert.Nil(t, err)

	p := make([]byte, 4)
	n, err := reader.Read(p)
	assert.Nil(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte{9, 8, 7, 6}, p)
}
