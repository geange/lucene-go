package selector

import (
	"encoding/binary"
	"math/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	values100   = make([][]byte, 0)
	values1000  = make([][]byte, 0)
	values10000 = make([][]byte, 0)
)

func init() {
	values100 = initValues(100)
	values1000 = initValues(1000)
	values10000 = initValues(10000)
}

func initValues(n int) [][]byte {
	res := make([][]byte, 0, n)

	for i := 0; i < n; i++ {
		x := rand.Uint32()
		y := rand.Uint32()
		z := rand.Uint32()
		bs := make([]byte, 4*3)
		binary.BigEndian.PutUint32(bs[0:], x)
		binary.BigEndian.PutUint32(bs[4:], y)
		binary.BigEndian.PutUint32(bs[8:], z)

		res = append(res, bs)
	}
	return res
}

func cloneValues(values [][]byte) [][]byte {
	res := make([][]byte, 0, len(values))
	for _, value := range values {
		bs := make([]byte, len(value))
		copy(bs, value)
		res = append(res, bs)
	}
	return res
}

func Test_RadixSelector_SelectK(t *testing.T) {
	doTestSelectK(t, 100, 5)
	doTestSelectK(t, 1000, 5)
	doTestSelectK(t, 10000, 5)
	doTestSelectK(t, 100, 50)
	doTestSelectK(t, 1000, 50)
	doTestSelectK(t, 10000, 50)
}

func doTestSelectK(t *testing.T, n, k int) {
	nums := make([]uint32, 0)

	radix := NewMockRadix()
	for i := 0; i < n; i++ {
		bs := make([]byte, 4)
		num := rand.Uint32()
		nums = append(nums, num)
		binary.BigEndian.PutUint32(bs, num)
		radix.Add(bs)
	}

	slices.Sort(nums)

	selector := NewRadixSelector(radix, 4)

	selector.SelectK(0, n, k)

	radixNums := make([]uint32, 0)
	for i := 0; i < k; i++ {
		num := binary.BigEndian.Uint32(radix.values[i])
		radixNums = append(radixNums, num)
	}
	slices.Sort(radixNums)

	assert.Equal(t, nums[:k], radixNums)
}

func Test_RadixSelector_SameValues(t *testing.T) {
	//nums := make([]uint32, 100)
	nBytes := make([][]byte, 100)

	for i := 0; i < 100; i++ {
		//nums[i] = 42
		nBytes[i] = binary.BigEndian.AppendUint32(nil, 42)
	}

	radix := NewMockRadix(nBytes...)
	NewRadixSelector(radix, 4).SelectK(0, 100, 10)
}

func Test_RadixSelector_Prefix(t *testing.T) {
	doTestPrefix(t, 50, 2)
	doTestPrefix(t, 100, 2)
	doTestPrefix(t, 101, 2)
	doTestPrefix(t, 1000, 2)
	doTestPrefix(t, 1000, 5)
}

func doTestPrefix(t *testing.T, n, k int) {
	nBytes := make([][]byte, n)

	for i := 0; i < n; i++ {
		nBytes[i] = binary.BigEndian.AppendUint32(nil, 42)
	}

	for i := 0; i < k; i++ {
		nBytes[n-i-1] = binary.BigEndian.AppendUint32(nil, 1)
	}

	radix := NewMockRadix(nBytes...)
	NewRadixSelector(radix, 4).SelectK(0, n, k+1)

	i := 0
	for ; i < k; i++ {
		assert.Equal(t, binary.BigEndian.Uint32(radix.values[i]), uint32(1))
	}
	assert.Equal(t, binary.BigEndian.Uint32(radix.values[i]), uint32(42))
}

func Benchmark_IntroSelector_Select5_100(b *testing.B) {
	n := 100
	k := 5

	for i := 0; i < b.N; i++ {
		values := cloneValues(values100)
		radix := NewMockRadix(values...)
		NewIntroSelector(radix).SelectK(0, n, k)
	}
}

func Benchmark_RadixSelector_Select5_100(b *testing.B) {
	length := 12
	n := 100
	k := 5

	for i := 0; i < b.N; i++ {
		values := cloneValues(values100)
		radix := NewMockRadix(values...)

		NewRadixSelector(radix, length).SelectK(0, n, k)
	}
}

func Benchmark_IntroSelector_Select5_1000(b *testing.B) {
	n := 1000
	k := 5

	for i := 0; i < b.N; i++ {
		values := cloneValues(values1000)
		radix := NewMockRadix(values...)

		NewIntroSelector(radix).SelectK(0, n, k)
	}
}

func Benchmark_RadixSelector_Select5_1000(b *testing.B) {
	length := 12
	n := 1000
	k := 5

	for i := 0; i < b.N; i++ {
		values := cloneValues(values1000)
		radix := NewMockRadix(values...)

		NewRadixSelector(radix, length).SelectK(0, n, k)
	}
}

func Benchmark_IntroSelector_Select5_10000(b *testing.B) {
	n := 10000
	k := 5

	for i := 0; i < b.N; i++ {
		values := cloneValues(values10000)
		radix := NewMockRadix(values...)

		NewIntroSelector(radix).SelectK(0, n, k)
	}
}

func Benchmark_RadixSelector_Select5_10000(b *testing.B) {
	length := 12
	n := 10000
	k := 5

	for i := 0; i < b.N; i++ {
		values := cloneValues(values10000)
		radix := NewMockRadix(values...)

		NewRadixSelector(radix, length).SelectK(0, n, k)
	}
}

func Benchmark_IntroSelector_Select50_100(b *testing.B) {
	n := 100
	k := 50

	for i := 0; i < b.N; i++ {
		values := cloneValues(values100)
		radix := NewMockRadix(values...)
		NewIntroSelector(radix).SelectK(0, n, k)
	}
}

func Benchmark_RadixSelector_Select50_100(b *testing.B) {
	length := 12
	n := 100
	k := 50

	for i := 0; i < b.N; i++ {
		values := cloneValues(values100)
		radix := NewMockRadix(values...)

		NewRadixSelector(radix, length).SelectK(0, n, k)
	}
}

func Benchmark_IntroSelector_Select50_1000(b *testing.B) {
	n := 1000
	k := 50

	for i := 0; i < b.N; i++ {
		values := cloneValues(values1000)
		radix := NewMockRadix(values...)

		NewIntroSelector(radix).SelectK(0, n, k)
	}
}

func Benchmark_RadixSelector_Select50_1000(b *testing.B) {
	length := 12
	n := 1000
	k := 50

	for i := 0; i < b.N; i++ {
		values := cloneValues(values1000)
		radix := NewMockRadix(values...)

		NewRadixSelector(radix, length).SelectK(0, n, k)
	}
}

func Benchmark_IntroSelector_Select50_10000(b *testing.B) {
	n := 10000
	k := 50

	for i := 0; i < b.N; i++ {
		values := cloneValues(values10000)
		radix := NewMockRadix(values...)

		NewIntroSelector(radix).SelectK(0, n, k)
	}
}

func Benchmark_RadixSelector_Select50_10000(b *testing.B) {
	length := 12
	n := 10000
	k := 50

	for i := 0; i < b.N; i++ {
		values := cloneValues(values10000)
		radix := NewMockRadix(values...)

		NewRadixSelector(radix, length).SelectK(0, n, k)
	}
}
