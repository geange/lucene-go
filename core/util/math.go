package util

import (
	"golang.org/x/exp/constraints"
	"math"
)

type Number interface {
	constraints.Integer | constraints.Float
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

func NextUp(f float64) float64 {
	if math.IsInf(f, 1) || math.IsNaN(f) {
		return f
	}
	bits := math.Float64bits(f)
	if bits&(1<<63) != 0 {
		bits--
	} else {
		bits++
	}
	return math.Float64frombits(bits)
}
