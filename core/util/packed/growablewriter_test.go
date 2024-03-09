package packed

import (
	"github.com/stretchr/testify/assert"
	"slices"
	"testing"
)

func TestNewGrowableWriter(t *testing.T) {
	startBitsPerValue := 1
	valueCount := 100
	acceptableOverheadRatio := 1.0
	growableWriter := NewGrowableWriter(startBitsPerValue, valueCount, acceptableOverheadRatio)
	growableWriter.Set(0, 1<<8-1)

	{
		n, err := growableWriter.Get(0)
		assert.Nil(t, err)
		assert.EqualValues(t, 1<<8-1, n)
	}

	assert.EqualValues(t, 8, growableWriter.GetBitsPerValue())
}

func TestGrowableWriter_Fill(t *testing.T) {
	startBitsPerValue := 1
	valueCount := 100
	acceptableOverheadRatio := 1.0
	direct := NewGrowableWriter(startBitsPerValue, valueCount, acceptableOverheadRatio)
	direct.Clear()

	direct.Fill(0, 10, 1)
	assert.EqualValues(t, 1, direct.GetTest(9))

	direct.Fill(0, 20, 2)
	assert.EqualValues(t, 2, direct.GetTest(9))
	assert.EqualValues(t, 2, direct.GetTest(17))

	direct.Fill(90, 100, 3)
	assert.EqualValues(t, 3, direct.GetTest(90))
	assert.EqualValues(t, 3, direct.GetTest(99))
}

func TestGrowableWriter_Set(t *testing.T) {
	startBitsPerValue := 1
	valueCount := 100
	acceptableOverheadRatio := 1.0
	direct := NewGrowableWriter(startBitsPerValue, valueCount, acceptableOverheadRatio)

	direct.Set(0, 2)
	assert.EqualValues(t, 2, direct.GetTest(0))

	direct.Set(1, 3)
	assert.EqualValues(t, 3, direct.GetTest(1))

	direct.Set(2, 4)
	assert.EqualValues(t, 4, direct.GetTest(2))

	direct.Set(99, 5)
	assert.EqualValues(t, 5, direct.GetTest(99))
}

func TestGrowableWriter_GetBulk(t *testing.T) {
	startBitsPerValue := 1
	valueCount := 1000
	acceptableOverheadRatio := 1.0
	direct := NewGrowableWriter(startBitsPerValue, valueCount, acceptableOverheadRatio)
	direct.Fill(0, 10, 1)

	bulk := make([]uint64, 10)
	direct.GetBulk(0, bulk)

	assert.EqualValues(t, fill(make([]uint64, 10), 1), bulk)

	direct.GetBulk(50, bulk)
	assert.EqualValues(t, make([]uint64, 10), bulk)
}

func TestGrowableWriter_SetBulk(t *testing.T) {
	startBitsPerValue := 1
	valueCount := 1000
	acceptableOverheadRatio := 1.0
	direct := NewGrowableWriter(startBitsPerValue, valueCount, acceptableOverheadRatio)

	index := 2

	writeBulk := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	expectBulk := slices.Clone(writeBulk)
	direct.SetBulk(index, writeBulk)

	readBulk := make([]uint64, 10)
	direct.GetBulk(index, readBulk)
	assert.EqualValues(t, expectBulk, readBulk)
}
