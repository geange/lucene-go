package util

import (
	"math"
)

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

// SumRelativeErrorBound
// Return a relative error bound for a sum of numValues positive doubles,
// computed using recursive summation, ie. sum = x1 + ... + xn.
// NOTE: This only works if all values are POSITIVE so that Σ |xi| == |Σ xi|.
// This uses formula 3.5 from Higham, Nicholas J. (1993),
// "The accuracy of floating point summation",
// SIAM Journal on Scientific Computing.
func SumRelativeErrorBound(numValues int) float64 {
	if numValues <= 1 {
		return 0
	}

	// u = unit roundoff in the paper, also called machine precision or machine epsilon
	u := 1.0 * math.Pow(2, -52) // 1 x 2^(-52)
	return float64(numValues-1) * u
}

func HighestOneBit(i int32) int32 {
	return i & (math.MinInt32 >> NumberOfLeadingZeros(i))
}

func NumberOfLeadingZeros(i int32) int {
	// HD, Count leading 0's
	if i <= 0 {
		if i == 0 {
			return 32
		}
		return 0
	}
	n := 31
	if i >= 1<<16 {
		n -= 16
		i >>= 16
	}
	if i >= 1<<8 {
		n -= 8
		i >>= 8
	}
	if i >= 1<<4 {
		n -= 4
		i >>= 4
	}
	if i >= 1<<2 {
		n -= 2
		i >>= 2
	}
	return n - int(i>>1)
}
