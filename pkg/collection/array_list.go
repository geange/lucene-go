package collection

import "errors"

type ArrayList[T any] struct {
	elem []T
}

func (r *ArrayList[T]) Size() int {
	return len(r.elem)
}

func (r *ArrayList[T]) Get(index int) (T, error) {
	if index < 0 || index >= r.Size() {
		return nil, ErrIndexOutOfBounds
	}
	return r.elem[index], nil
}

func (r *ArrayList[T]) Add(values T) error {
	if values == nil {
		return ErrNilPointer
	}
	r.elem = append(r.elem, values)
	return nil
}

func (r *ArrayList[T]) Set(index int, element T) error {
	if index < 0 || index >= r.Size() {
		return ErrIndexOutOfBounds
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
	ErrClassCast        = errors.New("class cast")          // if the class of the specified element prevents it from being added to this list
	ErrNilPointer       = errors.New("nil pointer")         // if the specified element is null and this list does not permit null elements
	ErrIndexOutOfBounds = errors.New("index out of bounds") // if the index is out of range (index < 0 || index >= size())

)
