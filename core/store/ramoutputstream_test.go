package store

import (
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRAMOutputStream(t *testing.T) {
	testRAMOutputStream(t, 64, 100)
	testRAMOutputStream(t, 64, 10)
	testRAMOutputStream(t, 1, 1)
	testRAMOutputStream(t, 200, 2080)
	testRAMOutputStream(t, 200, 3000)
}

func testRAMOutputStream(t *testing.T, outLoopSize, inLoopSize int) {
	directory := NewRAMDirectory()
	ramFile := NewRAMFile(directory)
	output := NewRAMOutputStream("test_001", ramFile, true)

	for i := 0; i < outLoopSize; i++ {
		n := i % 10

		char := '0' + byte(n)

		data := make([]byte, 0)
		for j := 0; j < inLoopSize; j++ {
			data = append(data, char)
		}
		_, err := output.Write(data)
		assert.Nil(t, err)
	}

	err := output.Close()
	assert.Nil(t, err)

	assert.Equal(t, output.GetFilePointer(), int64(outLoopSize*inLoopSize))

	next, stop := iter.Pull(ramFile.Iterator())
	defer stop()

	for i := 0; i < outLoopSize; i++ {
		n := i % 10

		for j := 0; j < inLoopSize; j++ {
			expect := '0' + byte(n)
			char, ok := next()
			assert.True(t, ok)
			assert.Equal(t, expect, char)
		}
	}
}
