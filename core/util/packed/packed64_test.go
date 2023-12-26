package packed

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPacked64_Fill(t *testing.T) {
	direct := NewPacked64(100, 10)
	direct.Clear()

	direct.Fill(0, 10, 1)
	assert.EqualValues(t, 1, direct.Get(9))

	direct.Fill(0, 20, 2)
	assert.EqualValues(t, 2, direct.Get(9))
	assert.EqualValues(t, 2, direct.Get(17))

	direct.Fill(90, 100, 3)
	assert.EqualValues(t, 3, direct.Get(90))
	assert.EqualValues(t, 3, direct.Get(99))
}

func TestPacked64_Set(t *testing.T) {
	direct := NewPacked64(100, 10)

	direct.Set(0, 2)
	assert.EqualValues(t, 2, direct.Get(0))

	direct.Set(1, 3)
	assert.EqualValues(t, 3, direct.Get(1))

	direct.Set(2, 4)
	assert.EqualValues(t, 4, direct.Get(2))

	direct.Set(99, 5)
	assert.EqualValues(t, 5, direct.Get(99))
}

func TestPacked64_GetBulk(t *testing.T) {
	direct := NewPacked64(100, 10)
	direct.Fill(0, 10, 1)

	bulk := make([]uint64, 10)
	direct.GetBulk(0, bulk)

	assert.EqualValues(t, fill(make([]uint64, 10), 1), bulk)

	direct.GetBulk(50, bulk)
	assert.EqualValues(t, make([]uint64, 10), bulk)
}

func TestPacked64_SetBulk(t *testing.T) {
	direct := NewPacked64(100, 10)

	index := 1

	writeBulk := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	expectBulk := slices.Clone(writeBulk)
	direct.SetBulk(index, writeBulk)

	readBulk := make([]uint64, 10)
	direct.GetBulk(index, readBulk)
	assert.EqualValues(t, expectBulk, readBulk)
}
