package structure

import "context"

type Iterator[T any] interface {
	HasNext() bool
	Next(context.Context) (T, error)
}

type Iterable[T any] interface {
	Iterator() Iterator[T]
}
