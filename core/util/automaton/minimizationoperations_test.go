package automaton

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestMinimize(t *testing.T) {
	t.Run("testIsTotal", func(t *testing.T) {
		a := NewAutomaton()
		init := a.CreateState()
		fini := a.CreateState()
		a.SetAccept(fini, true)
		err := a.AddTransition(init, fini, 0, unicode.MaxRune)
		assert.Nil(t, err)

		a.FinishState()
		assert.False(t, IsTotalAutomaton(a))
		err = a.AddTransition(fini, fini, 0, unicode.MaxRune)
		assert.Nil(t, err)

		a.FinishState()
		assert.False(t, IsTotalAutomaton(a))
		a.SetAccept(init, true)

		minimize, err := Minimize(a, DEFAULT_DETERMINIZE_WORK_LIMIT)
		assert.Nil(t, err)
		assert.True(t, IsTotalAutomaton(minimize))
	})

	t.Run("testMinimizeEmpty", func(t *testing.T) {
		a := NewAutomaton()
		init := a.CreateState()
		fini := a.CreateState()
		err := a.AddTransitionLabel(init, fini, 'a')
		assert.Nil(t, err)

		a.FinishState()
		a, err = Minimize(a, DEFAULT_DETERMINIZE_WORK_LIMIT)
		assert.Nil(t, err)

		assert.Equal(t, 0, a.GetNumStates())
	})

	t.Run("testMinus", func(t *testing.T) {
		t.Run("determinize", func(t *testing.T) {

			automata := NewAutomata()

			a1 := automata.MakeString("foobar")
			a2 := automata.MakeString("boobar")
			a3 := automata.MakeString("beebar")

			t.Run("test1", func(t *testing.T) {
				a, err := Union(a1, a2, a3)
				assert.Nil(t, err)

				a, err = determinize(a, DEFAULT_DETERMINIZE_WORK_LIMIT)
				assert.Nil(t, err)

				assertMatches(t, a, "foobar", "beebar", "boobar")
			})

			t.Run("test2", func(t *testing.T) {
				a, err := Union(a1, a2, a3)
				assert.Nil(t, err)

				minus, err := Minus(a, a2, DEFAULT_DETERMINIZE_WORK_LIMIT)
				assert.Nil(t, err)
				a4, err := determinize(minus, DEFAULT_DETERMINIZE_WORK_LIMIT)
				assert.Nil(t, err)
				assert.True(t, Run(a4, "foobar"))
				assert.False(t, Run(a4, "boobar"))
				assert.True(t, Run(a4, "beebar"))
				assertMatches(t, a4, "foobar", "beebar")
			})

			t.Run("test3", func(t *testing.T) {
				a, err := Union(a1, a2, a3)
				assert.Nil(t, err)

				minus, err := Minus(a, a2, DEFAULT_DETERMINIZE_WORK_LIMIT)
				assert.Nil(t, err)
				a4, err := determinize(minus, DEFAULT_DETERMINIZE_WORK_LIMIT)
				assert.Nil(t, err)

				minus, err = Minus(a4, a1, DEFAULT_DETERMINIZE_WORK_LIMIT)
				assert.Nil(t, err)
				a4, err = determinize(minus, DEFAULT_DETERMINIZE_WORK_LIMIT)
				assert.Nil(t, err)
				assert.False(t, Run(a4, "foobar"))
				assert.False(t, Run(a4, "boobar"))
				assert.True(t, Run(a4, "beebar"))
				assertMatches(t, a4, "beebar")
			})

			t.Run("test4", func(t *testing.T) {
				a, err := Union(a1, a2, a3)
				assert.Nil(t, err)

				minus, err := Minus(a, a2, DEFAULT_DETERMINIZE_WORK_LIMIT)
				assert.Nil(t, err)

				a4, err := determinize(minus, DEFAULT_DETERMINIZE_WORK_LIMIT)
				assert.Nil(t, err)

				minus, err = Minus(a4, a3, DEFAULT_DETERMINIZE_WORK_LIMIT)
				assert.Nil(t, err)

				a4, err = determinize(minus, DEFAULT_DETERMINIZE_WORK_LIMIT)
				assert.Nil(t, err)

				assert.True(t, Run(a4, "foobar"))
				assert.False(t, Run(a4, "boobar"))
				assert.False(t, Run(a4, "beebar"))
				assertMatches(t, a4, "foobar")
			})

			//

			//

		})
	})
}

func assertMatches(t *testing.T, automaton *Automaton, strings ...string) {
	expected := make(map[string]bool)
	for _, v := range strings {
		expected[v] = true
	}

	actual := make(map[string]bool)
	for _, v := range getFiniteStrings(automaton) {
		actual[v] = true
	}
	assert.EqualValues(t, expected, actual)
}

func getFiniteStrings(a *Automaton) []string {
	var result []string
	iter := NewFiniteStringsIterator(a, 0, a.GetNumStates()-1)

	for v := range iter.IteratorString() {
		result = append(result, v)
	}
	return result
}
