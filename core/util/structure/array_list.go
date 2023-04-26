package structure

import (
	"errors"
	"io"
)

type ArrayList[T any] struct {
	values []T
}

func NewArrayList[T any]() *ArrayList[T] {
	return NewArrayListArray([]T{})
}

func NewArrayListArray[T any](values []T) *ArrayList[T] {
	return &ArrayList[T]{values: values}
}
func (a *ArrayList[T]) Size() int {
	return len(a.values)
}

func (a *ArrayList[T]) ToArray() []T {
	return a.values
}

func (a *ArrayList[T]) Add(obj T) {
	a.values = append(a.values, obj)
}

func (a *ArrayList[T]) Set(idx int, obj T) error {
	if len(a.values) <= idx {
		return errors.New("out of index")
	}
	a.values[idx] = obj
	return nil
}

func (a *ArrayList[T]) Iterator() Iterator[T] {
	return &ArrayListIterator[T]{
		index:  0,
		values: a.values,
	}
}

type ArrayListIterator[T any] struct {
	none   T
	values []T
	index  int
}

func (a *ArrayListIterator[T]) HasNext() bool {
	return a.index < len(a.values)
}

func (a *ArrayListIterator[T]) Next() (T, error) {
	if a.index >= len(a.values) {
		return a.none, io.EOF
	}

	v := a.values[a.index]
	a.index++
	return v, nil
}
