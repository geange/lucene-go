package structure

type Iterator[T any] interface {
	HasNext() bool
	Next() (T, error)
}

type Iterable[T any] interface {
	Iterator() Iterator[T]
}
