package automaton

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRegExp(t *testing.T) {
	t.Run("testSmoke", func(t *testing.T) {
		r, err := NewRegExp("a(b+|c+)d")
		assert.Nil(t, err)

		automaton, err := r.ToAutomaton()
		assert.Nil(t, err)

		assert.True(t, Run(automaton, "abbbbbd"))
		assert.True(t, Run(automaton, "acd"))
		assert.False(t, Run(automaton, "ad"))
	})

	t.Run("testDeterminizeTooManyStates", func(t *testing.T) {
		r, err := NewRegExp("[ac]*a[ac]{50,200}")
		assert.Nil(t, err)
		_, err = r.ToAutomaton()
		assert.Error(t, err)
	})

	t.Run("testSerializeTooManyStatesToRepeat", func(t *testing.T) {
		r, err := NewRegExp("a{50001}")
		assert.Nil(t, err)
		_, err = r.toAutomaton(50000)
		assert.Error(t, err)
	})
}

//func TestNewRegExp(t *testing.T) {
//	regExp, err := NewRegExp("+-*(A|.....|BC)*]", WithSyntaxFlags(NONE))
//	assert.Nil(t, err)
//	fmt.Println(regExp)
//
//	automaton, err := regExp.ToAutomaton(1000000)
//	assert.Nil(t, err)
//
//	fmt.Println(automaton)
//}
