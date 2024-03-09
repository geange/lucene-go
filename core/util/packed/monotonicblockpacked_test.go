package packed

import (
	"context"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/geange/lucene-go/core/store"
	"github.com/stretchr/testify/assert"
)

func TestMonotonicBlockPacked(t *testing.T) {
	output := store.NewBufferDataOutput()
	writer := NewMonotonicBlockPackedWriter(output, 4096)

	valueCount := 100000

	nums := make([]uint64, 0)

	ctx := context.Background()
	for i := 0; i < valueCount; i++ {
		source := rand.NewSource(time.Now().UnixNano())
		n := rand.New(source).Intn(1<<20 - 1)
		nums = append(nums, uint64(n))
	}

	slices.Sort(nums)

	for _, v := range nums {
		err := writer.Add(ctx, v)
		assert.Nil(t, err)
	}

	err := writer.Finish(ctx)
	assert.Nil(t, err)

	bs := output.Bytes()

	input := store.NewBytesInput(bs)
	reader, err := NewMonotonicBlockPackedReader(ctx, input, VERSION_CURRENT, 4096, valueCount, false)
	assert.Nil(t, err)

	readNums := make([]uint64, 0)
	for i := 0; i < valueCount; i++ {
		n, err := reader.Get(i)
		assert.Nil(t, err)
		readNums = append(readNums, n)
	}

	assert.EqualValues(t, nums, readNums)
}
