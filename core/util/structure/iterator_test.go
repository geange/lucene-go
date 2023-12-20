package structure

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

var _ Iterator[int] = &iterator[int]{}

type iterator[T int] struct {
	i   int
	num []T
}

func (i *iterator[T]) HasNext() bool {
	return i.i < len(i.num)
}

func (i *iterator[T]) Next(ctx context.Context) (T, error) {
	if i.i >= len(i.num) {
		return 0, io.EOF
	}

	v := i.num[i.i]
	i.i++
	return v, nil
}

func TestIterator(t *testing.T) {
	it := iterator[int]{
		i:   0,
		num: []int{1, 2, 3, 4, 5},
	}

	v, err := it.Next(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 1, v)

	v, err = it.Next(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 2, v)

	v, err = it.Next(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 3, v)

	v, err = it.Next(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 4, v)

	v, err = it.Next(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 5, v)

	_, err = it.Next(context.TODO())
	assert.Error(t, err)
}
