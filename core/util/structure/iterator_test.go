package structure

import (
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

var _ Iterator[int] = &iterator{}

type iterator struct {
	i   int
	num []int
}

func (i *iterator) HasNext() bool {
	return i.i < len(i.num)
}

func (i *iterator) Next() (int, error) {
	if i.i >= len(i.num) {
		return 0, io.EOF
	}

	v := i.num[i.i]
	i.i++
	return v, nil
}

func TestIterator(t *testing.T) {
	it := iterator{
		i:   0,
		num: []int{1, 2, 3, 4, 5},
	}

	v, err := it.Next()
	assert.Nil(t, err)
	assert.Equal(t, 1, v)

	v, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, 2, v)

	v, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, 3, v)

	v, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, 4, v)

	v, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, 5, v)

	_, err = it.Next()
	assert.Error(t, err)
}
