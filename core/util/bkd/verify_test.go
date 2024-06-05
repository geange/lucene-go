package bkd

import (
	"bytes"
	"errors"
	"io"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/geange/lucene-go/core/store"
	"github.com/stretchr/testify/assert"
)

func getDirectory(numPoints int64) (store.Directory, error) {
	return store.NewNIOFSDirectory("test")
}

func verify(t *testing.T, config *Config, dir store.Directory,
	points PointWriter, start, end, middle int, sortedOnHeap int) {
	selector := NewRadixSelector(config, sortedOnHeap, dir, "test")

	//dataOnlyDims := config.numDims - config.numIndexDims

	// we only split by indexed dimension so we check for each only those dimension
	for splitDim := 0; splitDim < config.NumIndexDims(); splitDim++ {
		// We need to make a copy of the data as it is deleted in the process
		writer, err := copyPoints(config, dir, points)
		if !assert.Nil(t, err) {
			t.FailNow()
		}
		err = writer.Close()
		assert.Nil(t, err)

		inputSlice := NewPathSlice(writer, 0, points.Count())
		commonPrefixLengthInput, err := getRandomCommonPrefix(config, inputSlice, splitDim)
		assert.Nil(t, err)

		slices := make([]*PathSlice, 2)
		partitionPoint, err := selector.Select(inputSlice, slices, start, end, middle, splitDim, commonPrefixLengthInput)
		assert.Nil(t, err)
		assert.Equal(t, middle-start, slices[0].count)
		assert.Equal(t, end-middle, slices[1].count)

		err = slices[0].writer.Close()
		assert.Nil(t, err)
		err = slices[1].writer.Close()
		assert.Nil(t, err)

		// check that left and right slices contain the correct points
		maxDim, err := getMax(config, slices[0], splitDim)
		assert.Nil(t, err)
		minDim, err := getMin(config, slices[1], splitDim)
		assert.Nil(t, err)
		cmp := bytes.Compare(maxDim, minDim)
		assert.True(t, cmp <= 0)

		if cmp == 0 {
			maxDataDim, err := getMaxDataDimension(config, slices[0], maxDim, splitDim)
			assert.Nil(t, err)
			minDataDim, err := getMinDataDimension(config, slices[1], minDim, splitDim)
			assert.Nil(t, err)

			cmpCode := bytes.Compare(maxDataDim, minDataDim)
			assert.True(t, cmpCode <= 0)
			if cmpCode == 0 {
				maxDocID, err := getMaxDocId(config, slices[0], splitDim, partitionPoint, maxDataDim)
				assert.Nil(t, err)
				minDocId, err := getMinDocId(config, slices[1], splitDim, partitionPoint, minDataDim)
				assert.Nil(t, err)
				assert.True(t, minDocId >= maxDocID)
			}
		}
		assert.Equal(t, partitionPoint, minDim)
		err = slices[0].writer.Destroy()
		assert.Nil(t, err)
		err = slices[1].writer.Destroy()
		assert.Nil(t, err)
	}
}

func getMaxDocId(config *Config, p *PathSlice, dimension int, partitionPoint, dataDim []byte) (int, error) {
	docID := math.MinInt
	reader, err := p.writer.GetReader(p.start, p.count)
	if err != nil {
		return 0, err
	}

	offset := dimension * config.bytesPerDim
	dataOffset := config.packedIndexBytesLength
	dataLength := (config.numDims - config.numIndexDims) * config.bytesPerDim

	for {
		next, err := reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return 0, err
		}

		if !next {
			break
		}
		pointValue := reader.PointValue()
		packedValue := pointValue.PackedValue()

		dimValue := packedValue[offset : offset+config.BytesPerDim()]
		dataValue := packedValue[dataOffset : dataOffset+dataLength]

		if bytes.Compare(dimValue, partitionPoint[:config.BytesPerDim()]) == 0 &&
			bytes.Compare(dataValue, dataDim[:dataLength]) == 0 {

			newDocID := pointValue.DocID()
			if newDocID > docID {
				docID = newDocID
			}
		}
	}
	return docID, nil
}

func getMinDocId(config *Config, p *PathSlice, dimension int, partitionPoint, dataDim []byte) (int, error) {
	docID := math.MaxInt
	reader, err := p.writer.GetReader(p.start, p.count)
	if err != nil {
		return 0, err
	}

	offset := dimension * config.bytesPerDim
	dataOffset := config.packedIndexBytesLength
	dataLength := (config.numDims - config.numIndexDims) * config.bytesPerDim

	for {
		next, err := reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return 0, err
		}

		if !next {
			break
		}
		pointValue := reader.PointValue()
		packedValue := pointValue.PackedValue()

		dimValue := packedValue[offset : offset+config.BytesPerDim()]
		dataValue := packedValue[dataOffset : dataOffset+dataLength]

		if bytes.Compare(dimValue, partitionPoint[:config.BytesPerDim()]) == 0 &&
			bytes.Compare(dataValue, dataDim[:dataLength]) == 0 {

			newDocID := pointValue.DocID()
			if newDocID < docID {
				docID = newDocID
			}
		}
	}
	return docID, nil
}

func getMaxDataDimension(config *Config, p *PathSlice, maxDim []byte, splitDim int) ([]byte, error) {
	numDataDims := config.numDims - config.numIndexDims
	maxBytes := make([]byte, numDataDims*config.bytesPerDim)
	offset := splitDim * config.bytesPerDim

	reader, err := p.writer.GetReader(p.start, p.count)
	if err != nil {
		return nil, err
	}
	value := make([]byte, numDataDims*config.bytesPerDim)

	for {
		next, err := reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		if !next {
			break
		}

		packedValue := reader.PointValue().PackedValue()
		if bytes.Compare(maxDim[:config.bytesPerDim], packedValue[offset:offset+config.bytesPerDim]) == 0 {
			from := config.PackedIndexBytesLength()
			to := from + numDataDims*config.bytesPerDim
			copy(value, packedValue[from:to])

			if bytes.Compare(maxBytes, value) < 0 {
				copy(maxBytes, value)
			}
		}
	}
	return maxBytes, nil
}

func getMinDataDimension(config *Config, p *PathSlice, minDim []byte, splitDim int) ([]byte, error) {
	numDataDims := config.numDims - config.numIndexDims
	minBytes := make([]byte, numDataDims*config.bytesPerDim)

	for i := range minBytes {
		minBytes[i] = 0xFF
	}

	offset := splitDim * config.bytesPerDim

	reader, err := p.writer.GetReader(p.start, p.count)
	if err != nil {
		return nil, err
	}
	value := make([]byte, numDataDims*config.bytesPerDim)

	for {
		next, err := reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		if !next {
			break
		}

		packedValue := reader.PointValue().PackedValue()
		if bytes.Compare(minDim[:config.bytesPerDim], packedValue[offset:offset+config.bytesPerDim]) == 0 {
			from := config.PackedIndexBytesLength()
			to := from + numDataDims*config.bytesPerDim
			copy(value, packedValue[from:to])

			if bytes.Compare(minBytes, value) > 0 {
				copy(minBytes, value)
			}
		}
	}
	return minBytes, nil
}

// returns a common prefix length equal or lower than the current one
func getRandomCommonPrefix(config *Config, inputSlice *PathSlice, splitDim int) (int, error) {
	pointsMax, err := getMax(config, inputSlice, splitDim)
	if err != nil {
		return 0, err
	}
	pointsMin, err := getMin(config, inputSlice, splitDim)
	if err != nil {
		return 0, err
	}
	commonPrefixLength := Mismatch(pointsMin[:config.BytesPerDim()], pointsMax[:config.BytesPerDim()])
	if commonPrefixLength == -1 {
		commonPrefixLength = config.BytesPerDim()
	}
	source := rand.NewSource(time.Now().UnixNano())
	if nextBoolean(source) {
		return commonPrefixLength, nil
	}
	if commonPrefixLength == 0 {
		return 0, nil
	}
	return nextInt(source, 1, commonPrefixLength), nil
}

func getMin(config *Config, pathSlice *PathSlice, dimension int) ([]byte, error) {
	minBytes := make([]byte, config.BytesPerDim())
	for i := range minBytes {
		minBytes[i] = 0xFF
	}
	reader, err := pathSlice.writer.GetReader(pathSlice.start, pathSlice.count)
	if err != nil {
		return nil, err
	}

	value := make([]byte, config.BytesPerDim())

	for {
		next, err := reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if !next {
			break
		}

		pointValue := reader.PointValue()
		packedValue := pointValue.PackedValue()
		copy(value, packedValue[dimension*config.BytesPerDim():])

		if bytes.Compare(minBytes, value) > 0 {
			copy(minBytes, value)
		}
	}
	return minBytes, nil
}

func getMax(config *Config, pathSlice *PathSlice, dimension int) ([]byte, error) {
	maxBytes := make([]byte, config.BytesPerDim())

	reader, err := pathSlice.writer.GetReader(pathSlice.start, pathSlice.count)
	if err != nil {
		return nil, err
	}

	value := make([]byte, config.BytesPerDim())

	for {
		next, err := reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if !next {
			break
		}

		pointValue := reader.PointValue()
		packedValue := pointValue.PackedValue()
		copy(value, packedValue[dimension*config.BytesPerDim():])

		if bytes.Compare(maxBytes, value) < 0 {
			copy(maxBytes, value)
		}
	}
	return maxBytes, nil
}

func copyPoints(config *Config, dir store.Directory, points PointWriter) (PointWriter, error) {
	writer := getRandomPointWriter(config, dir, points.Count())
	err := points.Close()
	if err != nil {
		return nil, err
	}

	reader, err := points.GetReader(0, points.Count())
	if err != nil {
		return nil, err
	}
	for {
		next, err := reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
				break
			}
			return nil, err
		}
		if !next {
			break
		}

		pointValue := reader.PointValue()
		err = writer.AppendPoint(pointValue)
		if err != nil {
			return nil, err
		}
	}
	return writer, nil
}

func nextBoolean(source rand.Source) bool {
	return rand.New(source).Intn(100)%2 == 0
}
