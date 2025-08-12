package automaton

import "github.com/geange/lucene-go/core/util/array"

type IntsRefBuilder[T any] struct {
	ref *IntsRef[T]
}

func NewIntsRefBuilder[T any]() *IntsRefBuilder[T] {
	return &IntsRefBuilder[T]{ref: NewIntsRef[T]()}
}

type IntsRef[T any] struct {
	ints   []T
	offset int
	length int
}

func NewIntsRef[T any]() *IntsRef[T] {
	return &IntsRef[T]{
		ints: make([]T, 0),
	}
}

func (i *IntsRefBuilder[T]) SetLength(length int) {
	i.ref.length = length
}

func (i *IntsRefBuilder[T]) Length() int {
	return i.ref.length
}

func (i *IntsRefBuilder[T]) Clear() {
	i.SetLength(0)
}

func (i *IntsRefBuilder[T]) At(offset int) T {
	return i.ref.ints[offset]
}

func (i *IntsRefBuilder[T]) Set(offset int, value T) {
	i.ref.ints[offset] = value
}

func (i *IntsRefBuilder[T]) Append(value T) {
	if i.ref.offset+i.ref.length >= len(i.ref.ints) {
		i.ref.ints = append(i.ref.ints, value)
	}
	i.ref.length++
}

func (i *IntsRefBuilder[T]) Grow(depth int) {
	i.ref.ints = array.Grow(i.ref.ints, depth)
}

func (i *IntsRefBuilder[T]) Get() []T {
	return i.ref.ints
}
