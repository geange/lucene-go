package store

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferDataOutput(t *testing.T) {
	helloStr := "Hello world"

	srcOutput := NewBufferDataOutput()

	// simple write data
	n, err := srcOutput.Write([]byte(helloStr))
	assert.Nil(t, err)
	assert.Equal(t, len(helloStr), n)

	assert.Equal(t, []byte(helloStr), srcOutput.Bytes())

	// test reset
	srcOutput.Reset()
	assert.Equal(t, []byte{}, srcOutput.Bytes())

	// write multi time data
	n, err = srcOutput.Write([]byte(helloStr))
	assert.Nil(t, err)
	assert.Equal(t, len(helloStr), n)

	n, err = srcOutput.Write([]byte(helloStr))
	assert.Nil(t, err)
	assert.Equal(t, len(helloStr), n)

	assert.Equal(t, []byte(helloStr+helloStr), srcOutput.Bytes())

	// test CopyTo
	destOutput := NewBufferDataOutput()
	err = srcOutput.CopyTo(destOutput)
	assert.Nil(t, err)

	assert.Equal(t, []byte(helloStr+helloStr), destOutput.Bytes())

	n, err = destOutput.Write([]byte(helloStr))
	assert.Nil(t, err)
	assert.Equal(t, len(helloStr), n)
}

func TestBufferDataInput(t *testing.T) {
	buf := new(bytes.Buffer)
	buf.Write([]byte("1234567890"))
	input := NewBufferDataInput(buf)

	b1 := make([]byte, 1)
	n, err := input.Read(b1)
	assert.Nil(t, err)
	assert.Equal(t, 1, n)

	b10 := make([]byte, 10)
	n, err = input.Read(b10)
	assert.Nil(t, err)
	assert.Equal(t, 9, n)

	n, err = input.Read(b10)
	assert.NotNil(t, err)
	assert.Equal(t, 0, n)
}
