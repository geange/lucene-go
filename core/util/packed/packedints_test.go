package packed

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetMutable(t *testing.T) {
	m := GetMutable(100, 8, 0)
	for i := 0; i < 100; i++ {
		m.Set(i, int64(i+2))
	}

	for i := 0; i < 100; i++ {
		v := m.Get(i)
		assert.Equal(t, int64(i+2), v)
	}
}
