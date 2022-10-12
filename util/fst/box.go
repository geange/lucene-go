package fst

type Box[T any] struct {
	V T
}

func NewBox[T any](v T) *Box[T] {
	return &Box[T]{V: v}
}

func (*Box[T]) hashCode() int {
	panic("")
}
