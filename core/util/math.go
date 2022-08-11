package util

func Max[T int | int64](a, b T) T {
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

func Log(x, base int) int {
	ret := 0
	for x >= base {
		x /= base
		ret++
	}
	return ret
}
