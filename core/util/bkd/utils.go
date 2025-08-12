package bkd

import (
	"bytes"
	"errors"
	"math/rand"
	"sort"
	"time"

	"github.com/geange/lucene-go/core/store"
)

func sortN(from, to, k int, data sort.Interface) {
	for from < to {
		loc := sortPartition(from, to, data)
		if loc == k {
			return
		}
		if loc < k {
			sortN(loc+1, to, k, data)
			return
		}

		sortN(from, loc-1, k, data)
	}
}

func sortPartition(begin, end int, data sort.Interface) int {
	i, j := begin+1, end

	for i < j {
		if data.Less(begin, i) {
			data.Swap(i, j)
			j--
		} else {
			i++
		}
	}

	// 如果 values[begin] <= values[i]
	// !data.Less(begin, i) && !data.Less(i, begin) => values[begin] == values[i]
	if data.Less(begin, i) || (!data.Less(begin, i) && !data.Less(i, begin)) {
		i--
	}
	data.Swap(i, begin)
	return i
}

// Subtract Result = a - b, where a >= b, else IllegalArgumentException is thrown.
func Subtract(bytesPerDim, dim int, a, b, result []byte) error {
	start := dim * bytesPerDim
	end := start + bytesPerDim
	borrow := 0
	for i := end - 1; i >= start; i-- {
		diff := int(a[i]-b[i]) - borrow
		if diff < 0 {
			diff += 256
			borrow = 1
		} else {
			borrow = 0
		}
		result[i-start] = byte(diff)
	}
	if borrow != 0 {
		return errors.New("a < b")
	}
	return nil
}

func numberOfLeadingZeros(i int32) int {
	// HD, Count leading 0's
	if i <= 0 {
		if i == 0 {
			return 32
		}
		return 0
	}
	n := 31
	if i >= 1<<16 {
		n -= 16
		i >>= 16
	}
	if i >= 1<<8 {
		n -= 8
		i >>= 8
	}
	if i >= 1<<4 {
		n -= 4
		i >>= 4
	}
	if i >= 1<<2 {
		n -= 2
		i >>= 2
	}
	return n - int(i>>1)
}

func getRandomConfig() (*Config, error) {
	source := rand.NewSource(time.Now().UnixNano())
	numIndexDims := nextInt(source, 1, MAX_INDEX_DIMS)
	numDims := nextInt(source, numIndexDims, MAX_DIMS)
	bytesPerDim := nextInt(source, 2, 30)
	maxPointsInLeafNode := nextInt(source, 50, 2000)
	return NewConfig(numDims, numIndexDims, bytesPerDim, maxPointsInLeafNode)
}

func getOneDimConfig() (*Config, error) {
	source := rand.NewSource(time.Now().UnixNano())
	numIndexDims := 1
	numDims := 1
	bytesPerDim := nextInt(source, 2, 30)
	maxPointsInLeafNode := nextInt(source, 50, 2000)
	return NewConfig(numDims, numIndexDims, bytesPerDim, maxPointsInLeafNode)
}

func nextInt(source rand.Source, start, end int) int {
	return start + rand.New(source).Intn(end-start+1)
}

func nextBytes(source rand.Source, packedValue []byte) {
	rnd := rand.New(source)
	for i := range packedValue {
		packedValue[i] = byte(rnd.Intn(256))
	}
}

func getPackedValue(config *Config) []byte {
	source := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(source)
	packedValue := make([]byte, config.PackedBytesLength())
	for i := range packedValue {
		packedValue[i] = byte(rnd.Intn(256))
	}
	return packedValue
}

func getRandomPointWriter(config *Config, dir store.Directory, numPoints int) PointWriter {
	if numPoints < 4096 {
		return NewHeapPointWriter(config, int(numPoints))
	}
	return NewOfflinePointWriter(config, dir, "test", "data", numPoints)
}

func compareUnsigned(a []byte, aFromIndex, aToIndex int, b []byte, bFromIndex, bToIndex int) int {
	return bytes.Compare(a[aFromIndex:aToIndex], b[bFromIndex:bToIndex])
}

func arraycopy[T any](src []T, srcPos int, dest []T, destPos int, length int) {
	copy(dest[destPos:destPos+length], src[srcPos:srcPos+length])
}
