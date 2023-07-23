package numeric

import (
	"errors"
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

// LongToSortableBytes
// Encodes an long value such that unsigned byte order comparison is consistent with Long.compare(long, long)
// See Also: sortableBytesToLong(byte[], int)
func LongToSortableBytes(num int64, result []byte) {
	// Flip the sign bit so negative longs sort before positive longs:
	value := uint64(num)
	value ^= 0x8000000000000000
	result[0] = (byte)(value >> 56)
	result[1] = (byte)(value >> 48)
	result[2] = (byte)(value >> 40)
	result[3] = (byte)(value >> 32)
	result[4] = (byte)(value >> 24)
	result[5] = (byte)(value >> 16)
	result[6] = (byte)(value >> 8)
	result[7] = (byte)(value)
}

// SortableBytesToLong
// Decodes a long value previously written with longToSortableBytes
// See Also: longToSortableBytes(long, byte[], int)
func SortableBytesToLong(encoded []byte) int64 {
	v := (uint64(encoded[0]&0xFF) << 56) |
		(uint64(encoded[1]&0xFF) << 48) |
		(uint64(encoded[2]&0xFF) << 40) |
		(uint64(encoded[3]&0xFF) << 32) |
		(uint64(encoded[4]&0xFF) << 24) |
		(uint64(encoded[5]&0xFF) << 16) |
		(uint64(encoded[6]&0xFF) << 8) |
		uint64(encoded[7]&0xFF)
	// Flip the sign bit back
	v ^= 0x8000000000000000
	return int64(v)
}
