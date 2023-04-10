package util

type Number interface {
	int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64 | int | uint | float32 | float64
}

func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Min[T Number](a, b T) T {
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
