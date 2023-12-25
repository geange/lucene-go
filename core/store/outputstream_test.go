package store

import (
	"bytes"
	"hash/crc32"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ io.WriteCloser = &mockWriter{}

type mockWriter struct {
	*bytes.Buffer
}

func newMockWriter() *mockWriter {
	return &mockWriter{Buffer: new(bytes.Buffer)}
}

func (m *mockWriter) Close() error {
	return nil
}

func TestOutputStreamIndexOutput(t *testing.T) {
	ieee := crc32.NewIEEE()
	_, err := ieee.Write([]byte{1, 2, 3, 4})
	assert.Nil(t, err)

	w := newMockWriter()
	output := NewOutputStream("x", w)
	defer output.Close()

	err = output.WriteByte(1)
	assert.Nil(t, err)

	n, err := output.Write([]byte{2, 3, 4})
	assert.Nil(t, err)
	assert.EqualValues(t, 3, n)

	checksum, err := output.GetChecksum()
	assert.Nil(t, err)

	assert.Equal(t, ieee.Sum32(), checksum)

	pointer := output.GetFilePointer()
	assert.EqualValues(t, 4, pointer)
}
