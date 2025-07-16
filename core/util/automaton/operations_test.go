package automaton

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_concatenate(t *testing.T) {
	automata := NewAutomata()

	a1 := automata.MakeString("m")
	a2 := automata.MakeAnyString()
	a3 := automata.MakeString("n")
	a4 := automata.MakeAnyString()

	a, err := concatenate(a1, a2, a3, a4)
	assert.Nil(t, err)
	a, err = determinize(a, 10000)
	assert.Nil(t, err)

	if !assert.True(t, Run(a, "mn")) {
		t.Skip()
	}
	if !assert.True(t, Run(a, "mone")) {
		t.Skip()
	}
	if !assert.False(t, Run(a, "m")) {
		t.Skip()
	}
}
