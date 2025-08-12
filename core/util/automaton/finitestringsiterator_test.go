package automaton

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFiniteStringsIterator_Next(t *testing.T) {
	a, err := Union(NewAutomata().MakeString("dog"), NewAutomata().MakeString("duck"))
	assert.Nil(t, err)
	a, err = Minimize(a, DEFAULT_DETERMINIZE_WORK_LIMIT)
	assert.Nil(t, err)

	iterator := NewFiniteStringsIteratorBuilder(a).New()

	values := make([][]int, 0)
	for {
		value, err := iterator.Next()
		assert.Nil(t, err)
		if value == nil {
			break
		}
		values = append(values, value)
	}

	for _, v := range values {
		fmt.Println(v)
	}
}
