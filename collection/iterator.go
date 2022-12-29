package collection

type Iterator[T comparable] interface {
	HasNext() bool
	Next() T
	Remove() error
}
