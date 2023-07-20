//go:build go1.18 || go1.19 || go1.20

package util

import "github.com/geange/gods-generic/cmp"

func clear[K cmp.Ordered, V any](values map[K]V) {
	keys := make([]K, 0, len(values))
	for k := range values {
		v := k
		keys = append(keys, v)
	}

	for _, key := range keys {
		delete(values, key)
	}
}

func min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func max[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}
