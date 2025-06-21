package automaton

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getCommonPrefix(t *testing.T) {
	t.Run("testCommonPrefixEmpty", func(t *testing.T) {
		prefix, err := getCommonPrefix(defaultAutomata.MakeEmpty())
		assert.Nil(t, err)
		assert.Equal(t, "", prefix)
	})

	t.Run("testCommonPrefixEmptyString", func(t *testing.T) {
		prefix, err := getCommonPrefix(defaultAutomata.MakeEmptyString())
		assert.Nil(t, err)
		assert.Equal(t, "", prefix)
	})

	t.Run("testCommonPrefixAny", func(t *testing.T) {
		a, err := defaultAutomata.MakeAnyString()
		assert.Nil(t, err)
		prefix, err := getCommonPrefix(a)
		assert.Nil(t, err)
		assert.Equal(t, "", prefix)
	})

	t.Run("testCommonPrefixRange", func(t *testing.T) {
		a, err := defaultAutomata.MakeCharRange('a', 'b')
		assert.Nil(t, err)
		prefix, err := getCommonPrefix(a)
		assert.Nil(t, err)
		assert.Equal(t, "", prefix)
	})

	t.Run("testCommonPrefixTrailingKleenStar", func(t *testing.T) {
		a1, err := defaultAutomata.MakeString("foo")
		if !assert.Nil(t, err) {
			return
		}
		a2, err := defaultAutomata.MakeAnyString()
		if !assert.Nil(t, err) {
			return
		}
		a, err := concatenate(a1, a2)
		assert.Nil(t, err)
		prefix, err := getCommonPrefix(a)
		assert.Nil(t, err)
		assert.Equal(t, "foo", prefix)
	})

	t.Run("", func(t *testing.T) {
		a := NewAutomaton()
		init := a.CreateState()
		medial := a.CreateState()
		fini := a.CreateState()
		a.SetAccept(fini, true)
		err := a.AddTransitionLabel(init, medial, 'm')
		assert.Nil(t, err)
		err = a.AddTransitionLabel(init, fini, 'm')
		assert.Nil(t, err)
		err = a.AddTransitionLabel(medial, fini, 'o')
		assert.Nil(t, err)
		a.FinishState()

		prefix, err := getCommonPrefix(a)
		assert.Nil(t, err)
		assert.Equal(t, "m", prefix)
	})

}
