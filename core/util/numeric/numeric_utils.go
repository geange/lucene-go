package numeric

import (
	"encoding/binary"
	"errors"
	"math"
)

// Subtract Result = a - b, where a >= b, else IllegalArgumentException is thrown.
func Subtract(bytesPerDim, dim int, a, b, result []byte) error {
	start := dim * bytesPerDim
	end := start + bytesPerDim
	borrow := 0
	for i := end - 1; i >= start; i-- {
		diff := int(a[i]) - int(b[i]&0xff) - borrow
		if diff < 0 {
			diff += 256
			borrow = 1
		} else {
			borrow = 0
		}
		result[i-start] = byte(diff)
	}
	if borrow != 0 {
		return errors.New("a < b")
	}
	return nil
}

// Uint64ToSortableBytes
// Encodes an long value such that unsigned byte order comparison is consistent with Long.compare(long, long)
// See Also: sortableBytesToLong(byte[], int)
func Uint64ToSortableBytes(num uint64, result []byte) {
	// Flip the sign bit so negative longs sort before positive longs:
	value := num ^ 0x8000000000000000
	binary.BigEndian.PutUint64(result, value)
}

// SortableBytesToLong
// Decodes a long value previously written with Uint64ToSortableBytes
func SortableBytesToLong(encoded []byte) int64 {
	return int64(SortableBytesToUint64(encoded))
}

// SortableBytesToUint64
// Decodes a long value previously written with Uint64ToSortableBytes
func SortableBytesToUint64(encoded []byte) uint64 {
	num := binary.BigEndian.Uint64(encoded)
	num = num ^ 0x8000000000000000
	return num
}

// IntToSortableBytes
// Encodes an int32 value such that unsigned byte order comparison is consistent with Integer.compare(int, int)
// 请参阅: SortableBytesToInt(byte[], int)
func IntToSortableBytes(value int32, result []byte) {
	// Flip the sign bit, so negative ints sort before positive ints correctly:
	n := uint32(value) ^ 0x80000000
	binary.BigEndian.PutUint32(result, n)
}

// SortableBytesToInt
// Decodes an integer value previously written with intToSortableBytes
// 请参阅: IntToSortableBytes(int, byte[])
func SortableBytesToInt(encoded []byte) int32 {
	return int32(binary.BigEndian.Uint32(encoded))
}

// Float64ToSortableLong
// Converts a double value to a sortable signed long.
// The value is converted by getting their IEEE 754 floating-point "double format" bit layout and
// then some bits are swapped, to be able to compare the result as long. By this the precision is
// not reduced, but the value can easily used as a long. The sort order (including Double.NaN) is
// defined by Double.compareTo; NaN is greater than positive infinity.
// SortableLongToDouble
func Float64ToSortableLong(value float64) uint64 {
	return SortableFloat64Bits(math.Float64bits(value))
}

// SortableUint64ToFloat64
// Converts a sortable long back to a double.
// 请参阅: Float64ToSortableLong
func SortableUint64ToFloat64(encoded uint64) float64 {
	return math.Float64frombits(SortableFloat64Bits(encoded))
}

// SortableFloat64Bits
// Converts IEEE 754 representation of a double to sortable order (or back to the original)
func SortableFloat64Bits(bits uint64) uint64 {
	return bits ^ uint64(int64(bits)>>63)&0x7fffffffffffffff
}
