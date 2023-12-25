//go:build linux || darwin

package store

import (
	"os"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMmapDataInput(t *testing.T) {
	file, err := os.CreateTemp("", "")
	assert.Nil(t, err)
	defer os.Remove(file.Name())

	n, err := file.Write([]byte{1, 2, 3, 4, 5, 6})
	assert.Nil(t, err)
	assert.EqualValues(t, 6, n)
	file.Close()

	input, err := NewMmapDataInput(file.Name())
	assert.Nil(t, err)
	defer input.Close()

	assert.EqualValues(t, 6, input.reader.Len())

	bs := make([]byte, 8)
	n, err = input.Read(bs)
	assert.Nil(t, err)
	assert.EqualValues(t, 6, n)
	assert.EqualValues(t, []byte{1, 2, 3, 4, 5, 6}, bs[:6])

	slices.Delete(bs, 0, len(bs))
	_, err = input.Read(bs)
	assert.NotNil(t, err)
}

func TestMmapDataInput_Clone(t *testing.T) {
	file, err := os.CreateTemp("", "")
	assert.Nil(t, err)
	defer os.Remove(file.Name())

	n, err := file.Write([]byte{1, 2, 3, 4, 5, 6})
	assert.Nil(t, err)
	assert.EqualValues(t, 6, n)
	file.Close()

	input, err := NewMmapDataInput(file.Name())
	assert.Nil(t, err)
	defer input.Close()

	cloneInput := input.Clone().(*MmapDataInput)
	defer cloneInput.Close()

	assert.EqualValues(t, 6, input.reader.Len())

	bs := make([]byte, 2)
	n, err = input.Read(bs)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, n)
	assert.EqualValues(t, []byte{1, 2}, bs[:2])

	slices.Delete(bs, 0, len(bs))
	n, err = cloneInput.Read(bs)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, n)
	assert.EqualValues(t, []byte{1, 2}, bs[:2])
}
