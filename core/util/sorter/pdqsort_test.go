package sorter

import (
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPdqSorterSort(t *testing.T) {
	//doTestPdqSort(t, 100)
	//doTestPdqSort(t, 1000)
	//doTestPdqSort(t, 10000)
	doTestPdqSort(t, 100000)
	//doTestPdqSort(t, 1000000)
}

func doTestPdqSort(t *testing.T, size int) {
	values := make([]int, 0)

	for i := 0; i < size; i++ {
		n := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
		values = append(values, n)
	}

	expects := slices.Clone(values)

	mock := NewMockInt(values...)
	NewPdqSorter(mock).Sort(0, size)

	slices.Sort(expects)

	assert.Equal(t, expects, mock.values)
}
