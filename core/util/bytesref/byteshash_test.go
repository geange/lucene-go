package bytesref

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBytesHashAdd(t *testing.T) {
	checkBytesHashAdd(t, 1<<4, 2)
	checkBytesHashAdd(t, 1<<4, 16)
	checkBytesHashAdd(t, 1<<4, 1<<5)
	checkBytesHashAdd(t, 1<<4, 1<<3)
	checkBytesHashAdd(t, 1<<8, 1<<7)
	checkBytesHashAdd(t, 1<<8, 1<<10)
	checkBytesHashAdd(t, 1<<16, 1<<15)
}

func checkBytesHashAdd(t *testing.T, capacity int, size int) {
	sizeLists := []int{8, 128, 256}
	//sizeLists := []int{8}
	for _, byteSize := range sizeLists {
		t.Run(fmt.Sprintf("capacity=%d,valuesNum=%d,byteSize=%d", capacity, size, byteSize), func(t *testing.T) {
			bytesHash, err := NewBytesHash(
				NewBlockPool(GetAllocatorBuilder().NewDirect(BYTE_BLOCK_SIZE)),
				WithCapacity(capacity),
				WithStartArray(NewDirectStartArray(capacity)),
			)
			assert.Nil(t, err)

			bytesItems := make([][]byte, 0)
			bytesIdItems := make([]int, 0)

			for i := 0; i < size; i++ {
				bs := make([]byte, byteSize)
				nextBytes(rand.NewSource(time.Now().UnixNano()), bs)
				bytesItems = append(bytesItems, bs)
			}

			for _, bs := range bytesItems {
				bytesId, err := bytesHash.Add(bs)
				assert.Nil(t, err)
				bytesIdItems = append(bytesIdItems, bytesId)
			}

			for i, bytesID := range bytesIdItems {
				bs := bytesHash.Get(bytesID)
				assert.Equal(t, bs, bytesItems[i])
			}
		})
	}
}

func nextBytes(source rand.Source, packedValue []byte) {
	rnd := rand.New(source)
	i := 0
	for ; i < len(packedValue)-4; i += 4 {
		binary.BigEndian.AppendUint32(packedValue[i:], rnd.Uint32())
	}

	for i < len(packedValue) {
		packedValue[i] = byte(rnd.Intn(256))
		i++
	}
}
