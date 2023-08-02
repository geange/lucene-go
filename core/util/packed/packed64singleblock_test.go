package packed

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPacked64SingleBlock_Set(t *testing.T) {
	bits := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 16, 21, 32}

	for _, bit := range bits {
		block, err := NewPacked64SingleBlock(1000, bit)
		assert.Nil(t, err)

		for i := 0; i < 10; i++ {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			idx := r.Intn(1000)
			value := uint64(r.Intn(1 << bit))
			block.Set(idx, value)

			getNum, err := block.Get(idx)
			assert.Nil(t, err)
			assert.EqualValues(t, value, getNum)
		}
	}
}
