package packed

import (
	"context"
	"github.com/geange/lucene-go/core/store"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPacked64_Fill(t *testing.T) {
	for i := 1; i < 63; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 1 + r.Intn(200000)
		direct := NewPacked64(valueCount, i)
		direct.Clear()

		value := uint64(r.Intn(1 << i))
		direct.Fill(0, valueCount, value)

		for j := 0; j < 10; j++ {

			idx := r.Intn(valueCount)
			assert.EqualValuesf(t, value, direct.GetTest(idx), "bitsPerValue=%d", i)
		}

	}
}

func TestPacked64_Set(t *testing.T) {
	for i := 1; i < 63; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := 1 + r.Intn(200000)
		direct := NewPacked64(valueCount, i)
		direct.Clear()

		times := r.Intn(10000)
		for j := 0; j < times; j++ {
			idx := r.Intn(valueCount)

			value := uint64(r.Intn(1 << i))
			direct.Set(idx, value)

			assert.EqualValuesf(t, value, direct.GetTest(idx), "bitsPerValue=%d", i)
		}
	}
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

func TestPacked64_Save(t *testing.T) {
	for i := 0; i < 10; i++ {
		output := store.NewBufferDataOutput()

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		valueCount := r.Intn(10000)
		bitsPerValue := 1 + r.Intn(64)
		pack := NewPacked64(valueCount, bitsPerValue)

		err := pack.Save(context.TODO(), output)
		assert.Nil(t, err)
	}
}
