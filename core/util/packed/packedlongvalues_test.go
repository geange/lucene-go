package packed

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestPackedLongValues(t *testing.T) {
	shift := 6 + rand.Intn(15)
	acceptableOverheadRatio := 1.0
	longValuesBuilder := NewPackedLongValuesBuilder(1<<shift, acceptableOverheadRatio)

	nums := make([]int64, 0)

	size := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100000)

	for i := 0; i < size; i++ {
		v := int64(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100))
		err := longValuesBuilder.Add(v)
		nums = append(nums, v)
		assert.Nil(t, err)
	}

	longValues, err := longValuesBuilder.Build()
	assert.Nil(t, err)

	for i, num := range nums {
		v, err := longValues.Get(i)
		assert.Nil(t, err)
		assert.EqualValues(t, num, v)
	}

	iterator := longValues.Iterator()
	i := 0
	for iterator.HasNext() {
		v, err := iterator.Next()
		assert.Nil(t, err)
		assert.EqualValues(t, nums[i], v)
		i++
	}
}
