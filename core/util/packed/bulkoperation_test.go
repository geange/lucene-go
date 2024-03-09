package packed

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getMinBlockCount(blockPerBits, valuePerBits int) int {
	for blockCount := 1; blockCount <= 64; blockCount++ {
		if blockCount*blockPerBits%valuePerBits == 0 {
			return blockCount
		}
	}
	return 1
}

func testDecodeUint64(t *testing.T, blockBits, valueBits, maxBlockSize int, packed BulkOperation) {
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
		packed.EncodeUint64(values, encodeBlocks, iterations)

		decodeValues := make([]uint64, valueCount)
		packed.DecodeUint64(encodeBlocks, decodeValues, iterations)
		assert.EqualValues(t, values, decodeValues)
	}
}

func testDecodeBytes(t *testing.T, blockBits, valueBits, maxBlockSize int, packed BulkOperation) {
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
