package packed

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"slices"
	"testing"
	"time"
)

func TestNewGrowableWriter(t *testing.T) {
	for startBitsPerValue := 1; startBitsPerValue <= 64; startBitsPerValue++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 100 + r.Intn(10000)
		acceptableOverheadRatio := 1.0
		growableWriter := NewGrowableWriter(startBitsPerValue, valueCount, acceptableOverheadRatio)

		for i := 0; i < 10; i++ {
			idx := r.Intn(valueCount)
			value := uint64(1<<startBitsPerValue - 1)
			growableWriter.Set(idx, value)
			n, err := growableWriter.Get(idx)
			assert.Nil(t, err)
			assert.EqualValues(t, value, n)
		}
	}
}

func TestGrowableWriter_Fill(t *testing.T) {
	for startBitsPerValue := 1; startBitsPerValue <= 64; startBitsPerValue++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 100 + r.Intn(10000)
		acceptableOverheadRatio := 1.0
		growableWriter := NewGrowableWriter(startBitsPerValue, valueCount, acceptableOverheadRatio)

		for i := 0; i < 10; i++ {
			from := r.Intn(valueCount / 2)
			to := from + 1 + r.Intn(valueCount/2)

			value := uint64(1<<startBitsPerValue - 1)

			growableWriter.Fill(from, to, value)
			idx := from + r.Intn(to-from)

			assert.EqualValues(t, value, growableWriter.GetTest(idx))
		}
	}
}

func TestGrowableWriter_Set(t *testing.T) {
	for startBitsPerValue := 1; startBitsPerValue <= 64; startBitsPerValue++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 100 + r.Intn(10000)
		acceptableOverheadRatio := 1.0
		growableWriter := NewGrowableWriter(startBitsPerValue, valueCount, acceptableOverheadRatio)

		for i := 0; i < 10; i++ {
			idx := r.Intn(valueCount)
			value := uint64(1<<startBitsPerValue - 1)
			growableWriter.Set(idx, value)
			assert.EqualValues(t, value, growableWriter.GetTest(idx))
		}
	}
}

func TestGrowableWriter_GetBulk(t *testing.T) {
	for startBitsPerValue := 1; startBitsPerValue <= 64; startBitsPerValue++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 100 + r.Intn(10000)
		acceptableOverheadRatio := 1.0
		growableWriter := NewGrowableWriter(startBitsPerValue, valueCount, acceptableOverheadRatio)

		fromIndex := r.Intn(valueCount / 2)
		toIndex := fromIndex + 1 + r.Intn(valueCount/2)
		size := toIndex - fromIndex
		nums := make([]uint64, size)

		value := uint64(1<<startBitsPerValue - 1)
		for i := 0; i < 10; i++ {
			idx := r.Intn(size)
			nums[idx] = value
			growableWriter.Set(fromIndex+idx, value)
		}

		bulkNums := make([]uint64, size)
		bulkSize := growableWriter.GetBulk(fromIndex, bulkNums)
		assert.NotEqual(t, 0, bulkSize)
		assert.EqualValuesf(t, nums[:bulkSize], bulkNums[:bulkSize],
			"startBitsPerValue=%d,valueCount=%d,fromIndex=%d,toIndex=%d",
			startBitsPerValue, valueCount, fromIndex, toIndex)
	}
	//
	//startBitsPerValue := 1
	//valueCount := 1000
	//acceptableOverheadRatio := 1.0
	//direct := NewGrowableWriter(startBitsPerValue, valueCount, acceptableOverheadRatio)
	//direct.Fill(0, 10, 1)
	//
	//bulk := make([]uint64, 10)
	//direct.GetBulk(0, bulk)
	//
	//assert.EqualValues(t, fill(make([]uint64, 10), 1), bulk)
	//
	//direct.GetBulk(50, bulk)
	//assert.EqualValues(t, make([]uint64, 10), bulk)
}

func TestGrowableWriter_SetBulk(t *testing.T) {
	startBitsPerValue := 1
	valueCount := 1000
	acceptableOverheadRatio := 0.0
	direct := NewGrowableWriter(startBitsPerValue, valueCount, acceptableOverheadRatio)

	index := 2

	writeBulk := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	expectBulk := slices.Clone(writeBulk)
	direct.SetBulk(index, writeBulk)

	readBulk := make([]uint64, 10)
	direct.GetBulk(index, readBulk)
	assert.EqualValues(t, expectBulk, readBulk)
}
