package selector

import (
	"encoding/binary"
	"math/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntroSelectorSelectK(t *testing.T) {
	doTestIntroSelectK(t, 2, 1, false)
	doTestIntroSelectK(t, 100, 5, false)
	doTestIntroSelectK(t, 1000, 5, false)
	doTestIntroSelectK(t, 10000, 5, false)
	doTestIntroSelectK(t, 100000, 1, false)
	doTestIntroSelectK(t, 100000, 100, false)
	doTestIntroSelectK(t, 100000, 200, false)
}

func TestIntroSelectorSlowSelectK(t *testing.T) {
	doTestIntroSelectK(t, 2, 1, true)
	doTestIntroSelectK(t, 100, 5, true)
	doTestIntroSelectK(t, 1000, 5, true)
	doTestIntroSelectK(t, 10000, 5, true)
	doTestIntroSelectK(t, 100000, 1, true)
	doTestIntroSelectK(t, 100000, 100, true)
	doTestIntroSelectK(t, 100000, 200, true)
}

func doTestIntroSelectK(t *testing.T, n, k int, slow bool) {
	nums := make([]uint32, 0)
	values := make([][]byte, 0)

	for i := 0; i < n; i++ {
		num := rand.Uint32()
		bs := make([]byte, 4)
		nums = append(nums, num)
		binary.BigEndian.PutUint32(bs, num)
		values = append(values, bs)
	}

	radix := NewMockRadix(values...)

	slices.Sort(nums)

	if slow {
		NewIntroSelector(radix).(*introSelector).slowSelect(0, n, k)
	} else {
		NewIntroSelector(radix).SelectK(0, n, k)
	}

	radixNums := make([]uint32, 0)
	for i := 0; i < k; i++ {
		num := binary.BigEndian.Uint32(radix.values[i])
		radixNums = append(radixNums, num)
	}
	slices.Sort(radixNums)

	assert.Equal(t, nums[:k], radixNums)
}
