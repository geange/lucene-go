package core

var (
	ArrayUtil = &ArrayUtilImpl{}
)

type ArrayUtilImpl struct {
}

func Grow[T any](array []T, minSize int) []T {
	if len(array) < minSize {
		return append(array, make([]T, minSize-len(array))...)
	}
	return array
}
