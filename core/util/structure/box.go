package structure

type Box[T any] struct {
	value T
}

func NewBox[T any](v T) *Box[T] {
	return &Box[T]{value: v}
}

func (b *Box[T]) Value() T {
	return b.value
}
