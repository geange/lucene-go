package structure

import (
	"errors"
	"github.com/geange/lucene-go/util/fst"
)

type ArrayList[T any] struct {
	data []T
}

func NewArrayList[T any]() *ArrayList[T] {
	return &ArrayList[T]{
		data: make([]T, 0),
	}
}

func NewArrayListFromValues[T any](values []T) *ArrayList[T] {
	return &ArrayList[T]{
		data: values,
	}
}

func (a *ArrayList[T]) Get(idx int) (T, error) {
	if idx >= len(a.data) {
		return nil, errors.New("out of range")
	}
	return a.data[idx], nil
}

func (a *ArrayList[T]) Add(value T) {
	a.data = append(a.data, value)
}

func (a *ArrayList[T]) Size() int {
	return len(a.data)
}

func (a *ArrayList[T]) Clear(from, to int) error {
	if from < to && from >= 0 && to <= len(a.data) {
		a.data = append(a.data[:from], a.data[to:]...)
		return nil
	}
	return fst.ErrOutOfArrayRange
}

func (a *ArrayList[T]) Values() []T {
	return a.data
}
