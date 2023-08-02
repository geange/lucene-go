package packed

import (
	"context"
	"github.com/geange/lucene-go/core/store"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestBlockPackedWriter(t *testing.T) {
	for i := 0; i < 10; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		blockShift := 6 + r.Intn(12)
		valueCount := 10 + r.Intn(10000)
		testBlockPackedWriter(t, blockShift, valueCount, true)
		testBlockPackedWriter(t, blockShift, valueCount, false)
	}
}

func testBlockPackedWriter(t *testing.T, blockShift int, valueCount int, direct bool) {
	blockSize := 1 << blockShift

	output := store.NewBufferDataOutput()
	writer := NewBlockPackedWriter(output, blockSize)
	n := valueCount

	ctx := context.TODO()

	nums := make([]uint64, 0, n)
	for i := 0; i < n; i++ {
		v := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1 << 10)
		nums = append(nums, uint64(v))

		err := writer.Add(ctx, uint64(v))
		assert.Nil(t, err)
	}

	err := writer.Finish(ctx)
	assert.Nil(t, err)

	bs := output.Bytes()
	input := store.NewBytesInput(bs)

	reader, err := NewBlockPackedReader(ctx, input, VERSION_CURRENT, blockSize, n, direct)
	assert.Nil(t, err)

	getNums := make([]uint64, 0)
	for i := 0; i < n; i++ {
		num, err := reader.Get(i)
		assert.Nil(t, err)
		getNums = append(getNums, num)
	}

	assert.EqualValuesf(t, nums, getNums, "blockShift=%d valueCount=%d", blockShift, valueCount)
}

func TestBlockPackedReaderIterator(t *testing.T) {
	t.Run("Next", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			blockShift := 6 + r.Intn(12)
			valueCount := 10 + r.Intn(10000)
			testBlockPackedReaderIteratorNext(t, blockShift, valueCount)
		}
	})

	t.Run("Skip", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			blockShift := 6 + r.Intn(12)
			valueCount := 10 + r.Intn(10000)
			testBlockPackedReaderIteratorSkip(t, blockShift, valueCount)
		}
	})

	t.Run("NextSlices", func(t *testing.T) {
		t.Run("enough values, len(values) >= count", func(t *testing.T) {
			for i := 0; i < 10; i++ {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				blockShift := 6 + r.Intn(12)
				valueCount := 10 + r.Intn(10000)

				blockSize := 1 << blockShift

				output := store.NewBufferDataOutput()
				writer := NewBlockPackedWriter(output, blockSize)
				n := valueCount

				ctx := context.TODO()

				nums := make([]uint64, 0, n)
				for j := 0; j < n; j++ {
					v := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1 << 10)
					nums = append(nums, uint64(v))

					err := writer.Add(ctx, uint64(v))
					assert.Nil(t, err)
				}

				err := writer.Finish(ctx)
				assert.Nil(t, err)

				bs := output.Bytes()
				input := store.NewBytesInput(bs)

				iterator := NewBlockPackedReaderIterator(input, VERSION_CURRENT, blockSize, n)

				splitIndex := rand.Intn(min(len(iterator.values), len(nums)))

				slices, err := iterator.NextSlices(context.TODO(), splitIndex)
				assert.Nil(t, err)
				assert.EqualValues(t, nums[:splitIndex], slices)
			}
		})

		t.Run("less values, len(values) < count", func(t *testing.T) {
			for i := 0; i < 10; i++ {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				blockShift := 6 + r.Intn(12)
				valueCount := 10 + r.Intn(10000)

				blockSize := 1 << blockShift

				output := store.NewBufferDataOutput()
				writer := NewBlockPackedWriter(output, blockSize)
				n := valueCount

				ctx := context.TODO()

				nums := make([]uint64, 0, n)
				for j := 0; j < n; j++ {
					v := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1 << 10)
					nums = append(nums, uint64(v))

					err := writer.Add(ctx, uint64(v))
					assert.Nil(t, err)
				}

				err := writer.Finish(ctx)
				assert.Nil(t, err)

				bs := output.Bytes()
				input := store.NewBytesInput(bs)

				iterator := NewBlockPackedReaderIterator(input, VERSION_CURRENT, blockSize, n)

				err = iterator.Skip(context.TODO(), n-1)
				assert.Nil(t, err)

				slices, err := iterator.NextSlices(context.TODO(), 10)
				assert.Nil(t, err)
				assert.EqualValues(t, 1, len(slices))
			}
		})

	})
}

func testBlockPackedReaderIteratorNext(t *testing.T, blockShift int, valueCount int) {
	blockSize := 1 << blockShift

	output := store.NewBufferDataOutput()
	writer := NewBlockPackedWriter(output, blockSize)
	n := valueCount

	ctx := context.TODO()

	nums := make([]uint64, 0, n)
	for i := 0; i < n; i++ {
		v := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1 << 10)
		nums = append(nums, uint64(v))

		err := writer.Add(ctx, uint64(v))
		assert.Nil(t, err)
	}

	err := writer.Finish(ctx)
	assert.Nil(t, err)

	bs := output.Bytes()
	input := store.NewBytesInput(bs)

	iterator := NewBlockPackedReaderIterator(input, VERSION_CURRENT, blockSize, n)

	getNums := make([]uint64, 0)
	for i := 0; i < n; i++ {
		num, err := iterator.Next(context.TODO())
		assert.Nil(t, err)
		getNums = append(getNums, num)
	}

	assert.EqualValuesf(t, nums, getNums, "blockShift=%d valueCount=%d", blockShift, valueCount)
}

func testBlockPackedReaderIteratorSkip(t *testing.T, blockShift int, valueCount int) {
	blockSize := 1 << blockShift

	output := store.NewBufferDataOutput()
	writer := NewBlockPackedWriter(output, blockSize)
	n := valueCount

	ctx := context.TODO()

	nums := make([]uint64, 0, n)
	for i := 0; i < n; i++ {
		v := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1 << 10)
		nums = append(nums, uint64(v))

		err := writer.Add(ctx, uint64(v))
		assert.Nil(t, err)
	}

	err := writer.Finish(ctx)
	assert.Nil(t, err)

	bs := output.Bytes()
	input := store.NewBytesInput(bs)

	iterator := NewBlockPackedReaderIterator(input, VERSION_CURRENT, blockSize, n)

	getNums := make([]uint64, 0)

	skipNums := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(n)

	err = iterator.Skip(context.TODO(), skipNums)
	assert.Nil(t, err)

	for i := 0; i < n-skipNums; i++ {
		num, err := iterator.Next(context.TODO())
		assert.Nil(t, err)
		getNums = append(getNums, num)
	}

	assert.EqualValuesf(t, nums[skipNums:], getNums, "blockShift=%d valueCount=%d", blockShift, valueCount)
}

func testBlockPackedReaderIteratorNextSlices(t *testing.T, blockShift int, valueCount int) {
	blockSize := 1 << blockShift

	output := store.NewBufferDataOutput()
	writer := NewBlockPackedWriter(output, blockSize)
	n := valueCount

	ctx := context.TODO()

	nums := make([]uint64, 0, n)
	for i := 0; i < n; i++ {
		v := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1 << 10)
		nums = append(nums, uint64(v))

		err := writer.Add(ctx, uint64(v))
		assert.Nil(t, err)
	}

	err := writer.Finish(ctx)
	assert.Nil(t, err)

	bs := output.Bytes()
	input := store.NewBytesInput(bs)

	iterator := NewBlockPackedReaderIterator(input, VERSION_CURRENT, blockSize, n)

	err = iterator.Skip(context.TODO(), n-1)
	assert.Nil(t, err)

	slices, err := iterator.NextSlices(context.TODO(), 10)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(slices))

	//splitIndex := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(n)
	//
	//slices1, err := iterator.NextSlices(ctx, splitIndex)
	//assert.Nil(t, err)
	//assert.EqualValues(t, nums[:splitIndex], slices1)
	//
	//slices2, err := iterator.NextSlices(ctx, n-splitIndex)
	//assert.Nil(t, err)
	//assert.EqualValues(t, nums[splitIndex:], slices2)
}
