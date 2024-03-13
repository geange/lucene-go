package packed

import (
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDirect8_Fill(t *testing.T) {
	for i := 0; i < 100; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 2 + r.Intn(10000)
		direct := NewDirect8(valueCount)

		from := r.Intn(valueCount / 2)
		to := from + r.Intn(valueCount/2) + 1

		value := uint64(r.Intn(1 << 8))
		direct.Fill(from, to, value)

		for j := 0; j < 10; j++ {
			idx := from + r.Intn(to-from)
			assert.EqualValuesf(t, value, direct.GetTest(idx), "from=%d to=%d", from, to)
		}
	}
}

func TestDirect8_Set(t *testing.T) {
	direct := NewDirect8(100)

	direct.Set(0, 2)
	assert.EqualValues(t, 2, direct.GetTest(0))

	direct.Set(1, 3)
	assert.EqualValues(t, 3, direct.GetTest(1))

	direct.Set(2, 4)
	assert.EqualValues(t, 4, direct.GetTest(2))

	direct.Set(99, 5)
	assert.EqualValues(t, 5, direct.GetTest(99))
}

func TestDirect8_GetBulk(t *testing.T) {
	direct := NewDirect8(100)
	direct.Fill(0, 10, 1)

	bulk := make([]uint64, 10)
	direct.GetBulk(0, bulk)

	assert.EqualValues(t, fill(make([]uint64, 10), 1), bulk)

	direct.GetBulk(50, bulk)
	assert.EqualValues(t, make([]uint64, 10), bulk)
}

func TestDirect8_SetBulk(t *testing.T) {
	direct := NewDirect8(100)

	index := 1

	writeBulk := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	expectBulk := slices.Clone(writeBulk)
	direct.SetBulk(index, writeBulk)

	readBulk := make([]uint64, 10)
	direct.GetBulk(index, readBulk)
	assert.EqualValues(t, expectBulk, readBulk)
}

func fill[T any](array []T, v T) []T {
	for i := range array {
		array[i] = v
	}
	return array
}

func TestDirect16_Fill(t *testing.T) {
	for i := 0; i < 100; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 2 + r.Intn(10000)
		direct := NewDirect16(valueCount)

		from := r.Intn(valueCount / 2)
		to := from + 1 + r.Intn(valueCount/2)

		value := uint64(r.Intn(1 << 16))
		direct.Fill(from, to, value)

		for j := 0; j < 10; j++ {
			idx := from + r.Intn(to-from)
			assert.EqualValues(t, value, direct.GetTest(idx))
		}
	}
}

func TestDirect16_Set(t *testing.T) {
	direct := NewDirect16(100)

	direct.Set(0, 2)
	assert.EqualValues(t, 2, direct.GetTest(0))

	direct.Set(1, 3)
	assert.EqualValues(t, 3, direct.GetTest(1))

	direct.Set(2, 4)
	assert.EqualValues(t, 4, direct.GetTest(2))

	direct.Set(99, 5)
	assert.EqualValues(t, 5, direct.GetTest(99))
}

func TestDirect16_GetBulk(t *testing.T) {
	direct := NewDirect16(100)
	direct.Fill(0, 10, 1)

	bulk := make([]uint64, 10)
	direct.GetBulk(0, bulk)

	assert.EqualValues(t, fill(make([]uint64, 10), 1), bulk)

	direct.GetBulk(50, bulk)
	assert.EqualValues(t, make([]uint64, 10), bulk)
}

func TestDirect16_SetBulk(t *testing.T) {
	direct := NewDirect16(100)

	index := 1

	writeBulk := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	expectBulk := slices.Clone(writeBulk)
	direct.SetBulk(index, writeBulk)

	readBulk := make([]uint64, 10)
	direct.GetBulk(index, readBulk)
	assert.EqualValues(t, expectBulk, readBulk)
}

func TestDirect32_Fill(t *testing.T) {
	for i := 0; i < 100; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 2 + r.Intn(10000)
		direct := NewDirect32(valueCount)

		from := r.Intn(valueCount / 2)
		to := from + 1 + r.Intn(valueCount/2)

		value := uint64(r.Intn(1 << 32))
		direct.Fill(from, to, value)

		for j := 0; j < 10; j++ {
			idx := from + r.Intn(to-from)
			assert.EqualValues(t, value, direct.GetTest(idx))
		}
	}
}

func TestDirect32_Set(t *testing.T) {
	direct := NewDirect32(100)

	direct.Set(0, 2)
	assert.EqualValues(t, 2, direct.GetTest(0))

	direct.Set(1, 3)
	assert.EqualValues(t, 3, direct.GetTest(1))

	direct.Set(2, 4)
	assert.EqualValues(t, 4, direct.GetTest(2))

	direct.Set(99, 5)
	assert.EqualValues(t, 5, direct.GetTest(99))
}

func TestDirect32_GetBulk(t *testing.T) {
	direct := NewDirect32(100)
	direct.Fill(0, 10, 1)

	bulk := make([]uint64, 10)
	direct.GetBulk(0, bulk)

	assert.EqualValues(t, fill(make([]uint64, 10), 1), bulk)

	direct.GetBulk(50, bulk)
	assert.EqualValues(t, make([]uint64, 10), bulk)
}

func TestDirect32_SetBulk(t *testing.T) {
	direct := NewDirect32(100)

	index := 1

	writeBulk := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	expectBulk := slices.Clone(writeBulk)
	direct.SetBulk(index, writeBulk)

	readBulk := make([]uint64, 10)
	direct.GetBulk(index, readBulk)
	assert.EqualValues(t, expectBulk, readBulk)
}

func TestDirect64_Fill(t *testing.T) {
	for i := 0; i < 100; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 2 + r.Intn(10000)
		direct := NewDirect64(valueCount)

		from := r.Intn(valueCount / 2)
		to := from + 1 + r.Intn(valueCount/2)

		value := uint64(r.Int63())
		direct.Fill(from, to, value)

		for j := 0; j < 10; j++ {
			idx := from + r.Intn(to-from)
			assert.EqualValues(t, value, direct.GetTest(idx))
		}
	}
}

func TestDirect64_Set(t *testing.T) {
	direct := NewDirect64(100)

	direct.Set(0, 2)
	assert.EqualValues(t, 2, direct.GetTest(0))

	direct.Set(1, 3)
	assert.EqualValues(t, 3, direct.GetTest(1))

	direct.Set(2, 4)
	assert.EqualValues(t, 4, direct.GetTest(2))

	direct.Set(99, 5)
	assert.EqualValues(t, 5, direct.GetTest(99))
}

func TestDirect64_GetBulk(t *testing.T) {
	direct := NewDirect64(100)
	direct.Fill(0, 10, 1)

	bulk := make([]uint64, 10)
	direct.GetBulk(0, bulk)

	assert.EqualValues(t, fill(make([]uint64, 10), 1), bulk)

	direct.GetBulk(50, bulk)
	assert.EqualValues(t, make([]uint64, 10), bulk)
}

func TestDirect64_SetBulk(t *testing.T) {
	direct := NewDirect64(100)

	index := 1

	writeBulk := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	expectBulk := slices.Clone(writeBulk)
	direct.SetBulk(index, writeBulk)

	readBulk := make([]uint64, 10)
	direct.GetBulk(index, readBulk)
	assert.EqualValues(t, expectBulk, readBulk)
}
