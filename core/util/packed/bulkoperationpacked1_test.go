package packed

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBulkOperationPacked1_DecodeLongToLong(t *testing.T) {
	t.Run("first bit is 1", func(t *testing.T) {
		n := uint64(1) << 63
		src := []int64{int64(n)}
		dst := make([]int64, 64)
		packed1 := NewBulkOperationPacked1()
		packed1.DecodeInts(src, dst, 1)

		expect := make([]int64, 64)
		expect[0] = 1
		assert.EqualValues(t, expect, dst)
	})

	t.Run("last bit is 1", func(t *testing.T) {
		src := []int64{int64(1)}
		dst := make([]int64, 64)
		packed1 := NewBulkOperationPacked1()
		packed1.DecodeInts(src, dst, 1)

		expect := make([]int64, 64)
		expect[63] = 1
		assert.EqualValues(t, expect, dst)
	})

	t.Run("3 bit is 1", func(t *testing.T) {
		idx := 3
		n := uint64(1) << (64 - idx)
		src := []int64{int64(n)}
		dst := make([]int64, 64)
		packed1 := NewBulkOperationPacked1()
		packed1.DecodeInts(src, dst, 1)

		expect := make([]int64, 64)
		expect[2] = 1
		assert.EqualValues(t, expect, dst)
	})

	t.Run("3/65th bit is 1", func(t *testing.T) {
		idx1 := 3
		n1 := uint64(1) << (64 - idx1)
		idx2 := 1
		n2 := uint64(1) << (64 - idx2)
		src := []int64{int64(n1), int64(n2)}
		dst := make([]int64, 128)
		packed1 := NewBulkOperationPacked1()
		packed1.DecodeInts(src, dst, 2)

		expect := make([]int64, 128)
		expect[2] = 1
		expect[64] = 1
		assert.EqualValues(t, expect, dst)
	})
}

func TestBulkOperationPacked1_DecodeByteToLong(t *testing.T) {
	t.Run("first bit is 1", func(t *testing.T) {
		idx := 1
		n := byte(1) << (8 - idx)
		src := []byte{n}
		dst := make([]int64, 8)
		packed1 := NewBulkOperationPacked1()
		packed1.DecodeBytes(src, dst, 1)

		expect := make([]int64, 8)
		expect[0] = 1
		assert.EqualValues(t, expect, dst)
	})

	t.Run("last bit is 1", func(t *testing.T) {
		idx := 8
		src := []byte{1 << (8 - idx)}
		dst := make([]int64, 8)
		packed1 := NewBulkOperationPacked1()
		packed1.DecodeBytes(src, dst, 1)

		expect := make([]int64, 8)
		expect[7] = 1
		assert.EqualValues(t, expect, dst)
	})

	t.Run("3 bit is 1", func(t *testing.T) {
		idx := 3
		n := byte(1) << (8 - idx)
		src := []byte{n}
		dst := make([]int64, 8)
		packed1 := NewBulkOperationPacked1()
		packed1.DecodeBytes(src, dst, 1)

		expect := make([]int64, 8)
		expect[2] = 1
		assert.EqualValues(t, expect, dst)
	})

	t.Run("3/9th bit is 1", func(t *testing.T) {
		idx1 := 3
		n1 := byte(1) << (8 - idx1)
		idx2 := 1
		n2 := byte(1) << (8 - idx2)
		src := []byte{n1, n2}
		dst := make([]int64, 16)
		packed1 := NewBulkOperationPacked1()
		packed1.DecodeBytes(src, dst, 2)

		expect := make([]int64, 16)
		expect[2] = 1
		expect[8] = 1
		assert.EqualValues(t, expect, dst)
	})
}

func TestBulkOperationPacked1_DecodeByteToInt(t *testing.T) {

}
