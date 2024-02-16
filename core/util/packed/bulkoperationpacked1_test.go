package packed

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestBulkOperationPacked1_DecodeUint64(t *testing.T) {
	iterationTimes := 10
	blockPerBits := 64

	for iterations := 1; iterations < iterationTimes; iterations++ {
		valuesNum := blockPerBits * iterations

		values := make([]uint64, valuesNum)

		loopTimes := 200

		for i := 0; i < loopTimes; i++ {

			clear(values)

			oneSize := rand.Intn(valuesNum)
			for j := 0; j < oneSize; j++ {
				idx := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(valuesNum)
				values[idx] = 1
			}

			packed1 := NewBulkOperationPacked1()

			encodeBlocks := make([]uint64, iterations)
			packed1.EncodeUint64(values[:], encodeBlocks, iterations)

			decodeValues := make([]uint64, valuesNum)
			packed1.DecodeUint64(encodeBlocks, decodeValues, iterations)
			assert.EqualValues(t, values[:], decodeValues)
		}
	}
}

func TestBulkOperationPacked1_DecodeBytes(t *testing.T) {
	iterationTimes := 80
	blockPerBits := 8

	for iterations := 1; iterations < iterationTimes; iterations++ {
		valuesNum := blockPerBits * iterations

		values := make([]uint64, valuesNum)

		loopTimes := 200

		for i := 0; i < loopTimes; i++ {

			clear(values)

			oneSize := rand.Intn(valuesNum)
			for j := 0; j < oneSize; j++ {
				idx := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(valuesNum)
				values[idx] = 1
			}

			packed1 := NewBulkOperationPacked1()

			encodeBlocks := make([]byte, iterations)
			packed1.EncodeBytes(values[:], encodeBlocks, iterations)

			decodeValues := make([]uint64, valuesNum)
			packed1.DecodeBytes(encodeBlocks, decodeValues, iterations)
			assert.EqualValues(t, values[:], decodeValues)
		}
	}
}
