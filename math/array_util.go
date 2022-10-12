package math

func Grow[T any](values []T, miniSize int) []T {
	if len(values) >= miniSize {
		return values
	}

	size := miniSize - len(values)
	values = append(values, make([]T, size)...)
	return values
}
