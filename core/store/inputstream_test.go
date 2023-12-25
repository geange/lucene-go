package store

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestInputStreamDataInput(t *testing.T) {
	file, err := os.CreateTemp("", "")
	assert.Nil(t, err)

	defer func() {
		err = os.Remove(file.Name())
		if err != nil {
			t.Error(err)
		}
	}()

	n, err := file.Write([]byte{1, 2, 3})
	assert.Nil(t, err)
	assert.EqualValues(t, 3, n)

	_, err = file.Seek(0, 0)
	assert.Nil(t, err)

	assert.Nil(t, err)
	input := NewInputStream(file)
	defer input.Close()

	b1, err := input.ReadByte()
	assert.Nil(t, err)
	assert.EqualValues(t, 1, b1)

	bs := make([]byte, 2)
	n, err = input.Read(bs)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, n)
	assert.Equal(t, []byte{2, 3}, bs)

	_, err = input.Read(bs)
	assert.NotNil(t, err)

	_, err = input.Read(bs)
	assert.NotNil(t, err)
}
