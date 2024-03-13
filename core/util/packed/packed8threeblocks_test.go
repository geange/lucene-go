package packed

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPacked8ThreeBlocks_Fill(t *testing.T) {
	for i := 0; i < 10; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 100 + r.Intn(10000)

		direct := NewPacked8ThreeBlocks(valueCount)
		direct.Clear()

		fromIndex := r.Intn(valueCount / 2)
		toIndex := fromIndex + r.Intn(valueCount/2)

		value := uint64(r.Intn(1 << 24))

		direct.Fill(fromIndex, toIndex, value)

		for j := 0; j < 10; j++ {
			idx := fromIndex + r.Intn(toIndex-fromIndex)
			assert.EqualValues(t, value, direct.GetTest(idx))
		}
	}
}

func TestPacked8ThreeBlocks_Set(t *testing.T) {
	for i := 0; i < 10; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 100 + r.Intn(10000)

		direct := NewPacked8ThreeBlocks(valueCount)
		direct.Clear()

		for j := 0; j < 10; j++ {
			value := uint64(r.Intn(1 << 24))
			idx := r.Intn(valueCount)
			direct.Set(idx, value)
			assert.EqualValues(t, value, direct.GetTest(idx))
		}
	}
}

func TestPacked8ThreeBlocks_GetBulk(t *testing.T) {
	for i := 0; i < 10; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 100 + r.Intn(10000)

		direct := NewPacked8ThreeBlocks(valueCount)
		direct.Clear()

		fromIndex := r.Intn(valueCount / 2)
		toIndex := fromIndex + r.Intn(valueCount/2)

		size := toIndex - fromIndex
		nums := make([]uint64, size)

		for j := 0; j < 10; j++ {
			idx := r.Intn(size)
			value := uint64(r.Intn(1 << 24))
			nums[idx] = value
			direct.Set(fromIndex+idx, value)
		}

		bulkNums := make([]uint64, size)
		direct.GetBulk(fromIndex, bulkNums)

		assert.EqualValues(t, nums, bulkNums)
	}
}

func TestPacked8ThreeBlocks_SetBulk(t *testing.T) {
	for i := 0; i < 10; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 100 + r.Intn(10000)

		direct := NewPacked8ThreeBlocks(valueCount)
		direct.Clear()

		fromIndex := r.Intn(valueCount / 2)
		toIndex := fromIndex + r.Intn(valueCount/2)

		size := toIndex - fromIndex
		nums := make([]uint64, size)

		for j := 0; j < 10; j++ {
			idx := r.Intn(size)
			value := uint64(r.Intn(1 << 24))
			nums[idx] = value
		}

		direct.SetBulk(fromIndex, nums)

		bulkNums := make([]uint64, size)
		direct.GetBulk(fromIndex, bulkNums)

		assert.EqualValues(t, nums, bulkNums)
	}
}
