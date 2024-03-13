package bulkoperation

import (
	"math/rand"
	"testing"
	"time"

	"github.com/geange/lucene-go/core/util/packed/common"
	"github.com/stretchr/testify/assert"
)

func TestBulkOperationPacked_DecodeUint64(t *testing.T) {
	t.Run("bitsPerValue = 25", func(t *testing.T) {
		testDecodeUint64(t, 64, 25, 200, NewPacked(25))
	})

	t.Run("bitsPerValue = 26", func(t *testing.T) {
		testDecodeUint64(t, 64, 26, 200, NewPacked(26))
	})

	t.Run("bitsPerValue = 27", func(t *testing.T) {
		testDecodeUint64(t, 64, 27, 200, NewPacked(27))
	})

	t.Run("bitsPerValue = 28", func(t *testing.T) {
		testDecodeUint64(t, 64, 28, 200, NewPacked(28))
	})

	t.Run("bitsPerValue = 29", func(t *testing.T) {
		testDecodeUint64(t, 64, 29, 200, NewPacked(29))
	})

	t.Run("bitsPerValue = 30", func(t *testing.T) {
		testDecodeUint64(t, 64, 30, 200, NewPacked(30))
	})
}

func TestBulkOperationPacked_DecodeBytes(t *testing.T) {
	testDecodeBytes(t, 8, 25, 100, NewPacked(25))
}

func testDecodeUint64(t *testing.T, blockBits, valueBits, maxBlockSize int, operation common.BulkOperation) {
	minBlockCount := getMinBlockCount(blockBits, valueBits)

	for blockCount := 1; blockCount < maxBlockSize; blockCount++ {
		if blockCount*blockBits%valueBits != 0 {
			continue
		}

		valueCount := blockCount * blockBits / valueBits

		values := make([]uint64, valueCount)

		oneSize := rand.Intn(valueCount/2) + 1
		for j := 0; j < oneSize; j++ {
			tRand := rand.New(rand.NewSource(time.Now().UnixNano()))
			idx := tRand.Intn(valueCount)
			values[idx] = uint64(tRand.Intn(1<<valueBits - 1))
		}

		iterations := blockCount / minBlockCount

		encodeBlocks := make([]uint64, blockCount)
		operation.EncodeUint64(values, encodeBlocks, iterations)

		decodeValues := make([]uint64, valueCount)
		operation.DecodeUint64(encodeBlocks, decodeValues, iterations)
		assert.EqualValues(t, values, decodeValues)
	}
}

func testDecodeBytes(t *testing.T, blockBits, valueBits, maxBlockSize int, packed common.BulkOperation) {
	minBlockCount := getMinBlockCount(blockBits, valueBits)

	for blockCount := 1; blockCount < maxBlockSize; blockCount++ {
		if blockCount*blockBits%valueBits != 0 {
			continue
		}

		valueCount := blockCount * blockBits / valueBits

		values := make([]uint64, valueCount)

		oneSize := rand.Intn(valueCount/2+1) + 1
		for j := 0; j < oneSize; j++ {
			tRand := rand.New(rand.NewSource(time.Now().UnixNano()))
			idx := tRand.Intn(valueCount)
			values[idx] = uint64(tRand.Intn(1<<valueBits - 1))
		}

		iterations := blockCount / minBlockCount

		encodeBlocks := make([]byte, blockCount)
		packed.EncodeBytes(values, encodeBlocks, iterations)

		decodeValues := make([]uint64, valueCount)
		packed.DecodeBytes(encodeBlocks, decodeValues, iterations)
		assert.EqualValues(t, values, decodeValues)
	}
}

func getMinBlockCount(blockPerBits, valuePerBits int) int {
	for blockCount := 1; blockCount <= 64; blockCount++ {
		if blockCount*blockPerBits%valuePerBits == 0 {
			return blockCount
		}
	}
	return 1
}
