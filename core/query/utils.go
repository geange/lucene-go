package query

import "cmp"

func IsNaN[T cmp.Ordered](f T) bool {
	return f != f
}
