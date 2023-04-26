package structure

import (
	"fmt"
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

	for it.HasNext() {
		fmt.Println(it.Next())
	}
}
