package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesInput(t *testing.T) {
	input := NewBytesInput([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})

	v, err := input.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, byte(0), v)

	b10 := make([]byte, 10)
	n, err := input.Read(b10)
	assert.Nil(t, err)
	assert.Equal(t, 9, n)
}

func TestBytesOutput(t *testing.T) {
	bs := make([]byte, 10)
	output := NewBytesDataOutput(bs)
	n, err := output.Write([]byte("ab"))
	assert.Nil(t, err)
	assert.Equal(t, 2, n)

	assert.Equal(t, []byte("ab"), output.bytes[:2])
	assert.Equal(t, 2, output.pos)

	n, err = output.Write(make([]byte, 8))
	assert.Nil(t, err)
	assert.Equal(t, 8, n)
	assert.Equal(t, 10, output.pos)

	_, err = output.Write(make([]byte, 8))
	assert.NotNil(t, err)

}

func TestBytesOutputReset(t *testing.T) {
	bs := make([]byte, 10)
	output := NewBytesDataOutput(bs)
	n, err := output.Write([]byte("ab"))
	assert.Nil(t, err)
	assert.Equal(t, 2, n)

	err = output.Reset(make([]byte, 30))
	assert.Nil(t, err)

	n, err = output.Write(make([]byte, 30))
	assert.Nil(t, err)
	assert.Equal(t, 30, n)
	assert.Equal(t, 30, output.GetPosition())

	_, err = output.Write(make([]byte, 8))
	assert.NotNil(t, err)
}
