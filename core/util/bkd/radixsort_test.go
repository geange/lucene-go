package bkd

import (
	"bytes"
	"math/rand"
	"testing"
	"time"

	"github.com/geange/lucene-go/core/store"
	"github.com/stretchr/testify/assert"
)

func TestRandom(t *testing.T) {
	config, err := getRandomConfig()
	assert.Nil(t, err)

	source := rand.NewSource(time.Now().UnixNano())
	numPoints := nextInt(source, 1, DEFAULT_MAX_POINTS_IN_LEAF_NODE)
	points := NewHeapPointWriter(config, numPoints)
	value := make([]byte, config.packedBytesLength)
	for i := 0; i < numPoints; i++ {
		nextBytes(source, value)
		err := points.Append(value, i)
		assert.Nil(t, err)
	}
	verifySort(t, config, points, 0, numPoints)
}

func TestRandomAllEquals(t *testing.T) {
	config, err := getRandomConfig()
	assert.Nil(t, err)

	source := rand.NewSource(time.Now().UnixNano())
	numPoints := nextInt(source, 1, DEFAULT_MAX_POINTS_IN_LEAF_NODE)
	points := NewHeapPointWriter(config, numPoints)
	value := make([]byte, config.packedBytesLength)
	nextBytes(source, value)

	for i := 0; i < numPoints; i++ {
		err := points.Append(value, rand.New(source).Intn(numPoints))
		assert.Nil(t, err)
	}
	verifySort(t, config, points, 0, numPoints)
}

func TestRandomLastByteTwoValues(t *testing.T) {
	config, err := getRandomConfig()
	assert.Nil(t, err)

	source := rand.NewSource(time.Now().UnixNano())
	numPoints := nextInt(source, 1, DEFAULT_MAX_POINTS_IN_LEAF_NODE)
	points := NewHeapPointWriter(config, numPoints)
	value := make([]byte, config.packedBytesLength)
	nextBytes(source, value)

	for i := 0; i < numPoints; i++ {
		if rand.New(source).Intn(100)%2 == 0 {
			err := points.Append(value, 1)
			assert.Nil(t, err)
		} else {
			err := points.Append(value, 2)
			assert.Nil(t, err)
		}
	}
	verifySort(t, config, points, 0, numPoints)
}

func TestRandomFewDifferentValues(t *testing.T) {
	config, err := getRandomConfig()
	assert.Nil(t, err)

	source := rand.NewSource(time.Now().UnixNano())
	numPoints := nextInt(source, 1, DEFAULT_MAX_POINTS_IN_LEAF_NODE)
	points := NewHeapPointWriter(config, numPoints)

	numberValues := rand.New(source).Intn(8) + 2
	values := make([][]byte, numberValues)
	for i := 0; i < numberValues; i++ {
		values[i] = make([]byte, config.PackedBytesLength())
		nextBytes(source, values[i])
	}

	for i := 0; i < numPoints; i++ {
		err := points.Append(values[rand.Intn(numberValues)], i)
		assert.Nil(t, err)
	}
	verifySort(t, config, points, 0, numPoints)
}

func TestRandomDataDimDifferent(t *testing.T) {
	config, err := getRandomConfig()
	assert.Nil(t, err)

	source := rand.NewSource(time.Now().UnixNano())
	numPoints := nextInt(source, 1, DEFAULT_MAX_POINTS_IN_LEAF_NODE)
	points := NewHeapPointWriter(config, numPoints)
	value := make([]byte, config.PackedBytesLength())

	dataSize := (config.NumDims() - config.NumIndexDims()) * config.BytesPerDim()
	dataDimensionValues := make([]byte, dataSize)
	nextBytes(source, value)

	for i := 0; i < numPoints; i++ {
		source := rand.NewSource(time.Now().UnixNano())
		nextBytes(source, dataDimensionValues)
		from := config.PackedIndexBytesLength()
		to := from + dataSize
		copy(value[from:to], dataDimensionValues)
		err := points.Append(value, rand.New(source).Intn(numPoints))
		assert.Nil(t, err)
	}
	verifySort(t, config, points, 0, numPoints)
}

func verifySort(t *testing.T, config *Config, points *HeapPointWriter, start, end int) {
	dir, err := store.NewNIOFSDirectory("test")
	assert.Nil(t, err)
	defer dir.Close()

	radixSelector := NewRadixSelector(config, 1000, dir, "test")

	for splitDim := 0; splitDim < config.NumDims(); splitDim++ {
		commonPrefixLength := getHeapPointWriterRandomCommonPrefix(config, points, start, end, splitDim)
		radixSelector.HeapRadixSort(points, start, end, splitDim, commonPrefixLength)

		previous := make([]byte, config.PackedBytesLength())
		previousDocId := -1
		dimOffset := splitDim * config.BytesPerDim()
		for j := start; j < end; j++ {
			pointValue := points.GetPackedValueSlice(j)
			value := pointValue.PackedValue()
			from := dimOffset
			to := from + config.BytesPerDim()
			cmp := bytes.Compare(
				value[from:to],
				previous[from:to],
			)
			if !assert.True(t, cmp >= 0) {
				t.FailNow()
			}

			if cmp == 0 {
				dataOffset := config.NumIndexDims() * config.BytesPerDim()
				cmp = bytes.Compare(
					value[dataOffset:config.PackedBytesLength()],
					previous[dataOffset:config.PackedBytesLength()],
				)
				assert.True(t, cmp >= 0)
			}
			if cmp == 0 {
				if !assert.True(t, pointValue.DocID() >= previousDocId) {
					t.FailNow()
				}
			}
			copy(previous, value[:config.PackedBytesLength()])
			previousDocId = pointValue.DocID()
		}
	}
}

// returns a common prefix length equal or lower than the current one
func getHeapPointWriterRandomCommonPrefix(config *Config, points *HeapPointWriter, start, end, sortDim int) int {
	commonPrefixLength := config.bytesPerDim
	value := points.GetPackedValueSlice(start)
	bytesRef := value.PackedValue()
	firstValue := make([]byte, config.bytesPerDim)
	offset := sortDim * config.bytesPerDim
	copy(firstValue, bytesRef[offset:offset+config.bytesPerDim])

	for i := start + 1; i < end; i++ {
		value = points.GetPackedValueSlice(i)
		bytesRef = value.PackedValue()

		diff := Mismatch(
			bytesRef[offset:offset+config.bytesPerDim],
			firstValue[:config.bytesPerDim],
		)
		if diff != -1 && commonPrefixLength > diff {
			if diff == 0 {
				return diff
			}
			commonPrefixLength = diff
		}
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	flag := r.Intn(100)%2 == 1
	if flag {
		return commonPrefixLength
	}
	return r.Intn(commonPrefixLength) + 1
}
