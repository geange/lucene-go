package util

func Max[T int](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Min[T int](a, b T) T {
	if a > b {
		return b
	}
	return a
}
