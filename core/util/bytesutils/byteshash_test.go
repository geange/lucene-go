package bytesutils

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBytesHashAdd(t *testing.T) {
	checkBytesHashAdd(t, 16, 2)
	checkBytesHashAdd(t, 16, 16)
	checkBytesHashAdd(t, 1<<4, 1<<3)
	checkBytesHashAdd(t, 1<<8, 1<<7)
	checkBytesHashAdd(t, 1<<8, 1<<8)
	checkBytesHashAdd(t, 1<<16, 1<<15)
	checkBytesHashAdd(t, 1<<16, 1<<16)
}

func checkBytesHashAdd(t *testing.T, capacity int, size int) {
	bytesHash, err := NewBytesHash(
		NewBlockPool(NewDirectAllocator(BlockSize)),
		WithCapacity(capacity),
		WithStartArray(NewDirectBytesStartArray(capacity)),
	)
	assert.Nil(t, err)

	bytesItems := make([][]byte, 0)
	bytesIdItems := make([]int, 0)

	for i := 0; i < size; i++ {
		bs := make([]byte, 8)
		nextBytes(rand.NewSource(time.Now().UnixNano()), bs)

		bytesItems = append(bytesItems, bs)
		bytesId, err := bytesHash.Add(bs)
		assert.Nil(t, err)
		if assert.Truef(t, bytesId >= 0, "bytesId=%d", bytesId) {
			bytesIdItems = append(bytesIdItems, bytesId)
		}
	}

	for i, bytesID := range bytesIdItems {
		bs := bytesHash.Get(bytesID)
		assert.Equal(t, bs, bytesItems[i])
	}
}

func nextBytes(source rand.Source, packedValue []byte) {
	rnd := rand.New(source)
	for i := range packedValue {
		packedValue[i] = byte(rnd.Intn(256))
	}
}
