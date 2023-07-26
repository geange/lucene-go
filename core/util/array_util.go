package util

var (
	ArrayUtil = &ArrayUtilImpl{}
)

type ArrayUtilImpl struct {
}

func Oversize[T int | int64](minTargetSize, bytesPerElement T) T {
	if minTargetSize%4 != 0 {
		minTargetSize = (minTargetSize%bytesPerElement + 1) * bytesPerElement
	}
	return minTargetSize
}

func Grow[T any](array []T, minSize int) []T {
	if len(array) < minSize {
		return append(array, make([]T, minSize-len(array))...)
	}
	return array
}

func GrowExact[T any](array []T, size int) []T {
	newArray := make([]T, size)
	copy(newArray, array)
	return newArray
}

func Mismatch(a, b []byte) int {
	aLen, bLen := len(a), len(b)
	size := min(aLen, bLen)
	for i := 0; i < size; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	if aLen == bLen {
		return -1
	}
	return size
}
