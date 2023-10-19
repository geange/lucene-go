package bkd

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestRadixSelectorBasic(t *testing.T) {
	doTestRadixSelectorWithSize(t, 5)
	doTestRadixSelectorWithSize(t, 10)
	doTestRadixSelectorWithSize(t, 100)
	doTestRadixSelectorWithSize(t, 1000)
	doTestRadixSelectorWithSize(t, 4000)
}

func doTestRadixSelectorWithSize(t *testing.T, size int) {
	err := mkEmptyDir("test")
	assert.Nil(t, err)

	dir, err := getDirectory(int64(size))
	assert.Nil(t, err)
	defer dir.Close()

	config, err := getRandomConfig()
	assert.Nil(t, err)

	assert.Nil(t, err)
	points := getRandomPointWriter(config, dir, size)

	middle := func() int {
		num := rand.Intn(size)
		if num == 0 {
			return 1
		}
		return num
	}()

	for i := 0; i < size; i++ {
		err = points.Append(getPackedValue(config), i)
		assert.Nil(t, err)
	}
	err = points.Close()
	assert.Nil(t, err)
	pointWriter, err := copyPoints(config, dir, points)
	assert.Nil(t, err)
	verify(t, config, dir, pointWriter, 0, size, middle, 0)
}

func TestRadixSelectorOffline(t *testing.T) {
	size := 8192

	err := mkEmptyDir("test")
	assert.Nil(t, err)

	dir, err := getDirectory(int64(size))
	assert.Nil(t, err)
	defer dir.Close()

	config, err := NewConfig(1, 1, 4, 1024)
	assert.Nil(t, err)

	assert.Nil(t, err)
	points := getRandomPointWriter(config, dir, size)

	middle := 3

	value := make([]byte, 4)

	for i := 0; i < size; i++ {
		binary.BigEndian.PutUint32(value, uint32(i+1))
		err = points.Append(value, i)
		assert.Nil(t, err)
	}
	err = points.Close()
	assert.Nil(t, err)
	pointWriter, err := copyPoints(config, dir, points)
	assert.Nil(t, err)
	verify(t, config, dir, pointWriter, 0, size, middle, 1024)
}
