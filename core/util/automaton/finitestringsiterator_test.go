package automaton

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFiniteStringsIterator_Next(t *testing.T) {
	a1, err := defaultAutomata.MakeString("dog")
	assert.Nil(t, err)
	a2, err := defaultAutomata.MakeString("duck")
	assert.Nil(t, err)
	a, err := union(a1, a2)
	assert.Nil(t, err)
	ma, err := Minimize(a, DEFAULT_DETERMINIZE_WORK_LIMIT)
	assert.Nil(t, err)

	iterator := NewFiniteStringsIterator(ma, 0, -1)

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
