package collection

import "errors"

type ArrayList[T any] struct {
	elem []T
	no   T
}

func NewArrayList[T any]() *ArrayList[T] {
	return &ArrayList[T]{
		elem: make([]T, 0),
	}
}

func (r *ArrayList[T]) Size() int {
	return len(r.elem)
}

func (r *ArrayList[T]) Get(index int) (T, error) {
	if index < 0 || index >= r.Size() {
		return r.no, ErrIndexOutOfRange
	}
	return r.elem[index], nil
}

func (r *ArrayList[T]) Add(values T) error {
	r.elem = append(r.elem, values)
	return nil
}

func (r *ArrayList[T]) Set(index int, element T) error {
	if index < 0 || index >= r.Size() {
		return ErrIndexOutOfRange
	}
	r.elem[index] = element
	return nil
}

func (r *ArrayList[T]) ClearSubList(fromIndex, toIndex int) error {
	r.elem = append(r.elem[:fromIndex], r.elem[toIndex:]...)
	return nil
}

func (r *ArrayList[T]) List() []T {
	return r.elem
}

var (
	// ErrIndexOutOfRange if the index is out of range (index < 0 || index >= size())
	ErrIndexOutOfRange = errors.New("index out of range")
)
