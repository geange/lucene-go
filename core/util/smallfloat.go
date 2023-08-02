package util

import (
	"math"
	"math/bits"
)

var (
	MaxInt4       = LongToInt4(math.MaxInt32)
	NumFreeValues = 255 - MaxInt4
)

func IntToByte4(i int) byte {
	if i < NumFreeValues {
		return byte(i)
	} else {
		return (byte)(NumFreeValues + LongToInt4(int64(i-NumFreeValues)))
	}
}

func Int4ToLong(i int) int64 {
	b := i & 0x07
	shift := (i >> 3) - 1
	decoded := 0
	if shift == -1 {
		// subnormal value
		decoded = b
	} else {
		// normal value
		decoded = (b | 0x08) << shift
	}
	return int64(decoded)
}

func LongToInt4(i int64) int {
	numBits := 64 - bits.LeadingZeros64(uint64(i))
	if numBits < 4 {
		// subnormal value
		return int(i)
	} else {
		// normal value
		shift := int64(numBits - 4)
		// only keep the 5 most significant bits
		encoded := i >> shift
		// clear the most significant bit, which is implicit
		encoded &= 0x07
		// encode the shift, adding 1 because 0 is reserved for subnormal values
		encoded = encoded | ((shift + 1) << 3)
		return int(encoded)
	}
}

func Byte4ToInt(b byte) int {
	i := int(b)
	if i < NumFreeValues {
		return i
	} else {
		return NumFreeValues + int(Int4ToLong(i-NumFreeValues))
	}
}
