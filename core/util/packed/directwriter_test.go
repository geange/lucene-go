package packed

import (
	"encoding/binary"
	"github.com/geange/lucene-go/core/store"
	"github.com/stretchr/testify/assert"
	"math"
	"math/rand"
	"testing"
)

func TestNewDirectWriter(t *testing.T) {
	testDirectRW(t, 100, 1)
	testDirectRW(t, 100, 2)
	testDirectRW(t, 100, 4)
	testDirectRW(t, 100, 8)
	testDirectRW(t, 100, 12)
	testDirectRW(t, 100, 16)
	testDirectRW(t, 100, 20)
	testDirectRW(t, 100, 24)
	testDirectRW(t, 100, 28)
	testDirectRW(t, 100, 32)
	testDirectRW(t, 100, 40)
	testDirectRW(t, 100, 48)
	testDirectRW(t, 100, 56)
	testDirectRW(t, 100, 64)
}

func testDirectRW(t *testing.T, numValues, bitsPerValue int) {
	dataOutput := store.NewBufferDataOutput()

	directWriter, err := NewDirectWriter(dataOutput, numValues, bitsPerValue)
	assert.Nil(t, err)

	expectValues := make([]int, 0)

	maxNum := 1<<bitsPerValue - 1
	if maxNum < 0 {
		maxNum = math.MaxInt
	}

	for i := 0; i < numValues; i++ {
		n := rand.Intn(maxNum)
		expectValues = append(expectValues, n)
		err = directWriter.Add(uint64(n))
		assert.Nil(t, err)
	}

	err = directWriter.Finish()
	assert.Nil(t, err)

	accessInput := store.NewBytesRandomAccessInput(dataOutput.Bytes(), binary.BigEndian)

	reader, err := NewDirectReader().GetInstance(accessInput, bitsPerValue, 0)
	assert.Nil(t, err)

	for i := 0; i < numValues; i++ {
		v, err := reader.Get(int64(i))
		assert.Nil(t, err)
		assert.EqualValues(t, expectValues[i], v)
	}
}
