package sorter

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMSBRadixSorterSort(t *testing.T) {
	doTestMSBRadixSorter(t, 100)
	doTestMSBRadixSorter(t, 1000)
	doTestMSBRadixSorter(t, 10000)
	doTestMSBRadixSorter(t, 100000)
	doTestMSBRadixSorter(t, 500000)
}

func doTestMSBRadixSorter(t *testing.T, size int) {
	expects := make([][]byte, 0)
	actual := make([][]byte, 0)

	for i := 0; i < size; i++ {
		n := rand.New(rand.NewSource(time.Now().UnixNano())).Uint32()
		expects = append(expects, binary.BigEndian.AppendUint32(nil, n))
		actual = append(actual, binary.BigEndian.AppendUint32(nil, n))
	}

	slices.SortFunc(expects, bytes.Compare)

	radix := NewMock(actual...)
	NewMsbRadixSorter(size, radix).Sort(0, size)

	assert.Equal(t, expects, radix.values)
}
