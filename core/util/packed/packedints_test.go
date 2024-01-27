package packed

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetMutable(t *testing.T) {
	t.Run("num=100", func(t *testing.T) {
		num := 100

		m := GetMutable(num, 64, 0)
		for i := 0; i < num; i++ {
			m.Set(i, uint64(i+2))
		}

		for i := 0; i < num; i++ {
			v := m.Get(i)
			assert.EqualValues(t, i+2, v)
		}
	})

	t.Run("num=1000", func(t *testing.T) {
		num := 1000

		m := GetMutable(num, 64, 0)
		for i := 0; i < num; i++ {
			m.Set(i, uint64(i+2))
		}

		for i := 0; i < num; i++ {
			v := m.Get(i)
			assert.EqualValues(t, i+2, v)
		}
	})

	t.Run("num=10000", func(t *testing.T) {
		num := 10000

		m := GetMutable(num, 64, 0)
		for i := 0; i < num; i++ {
			m.Set(i, uint64(i+2))
		}

		for i := 0; i < num; i++ {
			v := m.Get(i)
			assert.EqualValues(t, i+2, v)
		}
	})

	t.Run("num=100000", func(t *testing.T) {
		num := 100000

		m := GetMutable(num, 64, 0)
		for i := 0; i < num; i++ {
			m.Set(i, uint64(i+2))
		}

		for i := 0; i < num; i++ {
			v := m.Get(i)
			assert.EqualValues(t, i+2, v)
		}
	})
}
