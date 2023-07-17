package store

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReaderX(t *testing.T) {
	source := new(bytes.Buffer)
	reader := NewReaderX(source)
	writer := NewWriterX(source)

	err := writer.WriteByte('a')
	assert.Nil(t, err)
	err = writer.WriteByte('b')
	assert.Nil(t, err)
	err = writer.WriteByte('c')
	assert.Nil(t, err)

	c1, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, byte('a'), c1)

	c2, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, byte('b'), c2)

	c3, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, byte('c'), c3)
}
