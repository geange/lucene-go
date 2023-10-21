package fst

import (
	"math/bits"
)

const (
	BYTE_SIZE    = 8
	INTEGER_SIZE = 32
	LONG_SIZE    = 64
	LONG_BYTES   = 8
)

// Returns whether the bit at given zero-based index is set. Example: bitIndex 10 means the third
// bit on the right of the second byte.
// Params:
//
//	bitIndex – The bit zero-based index. It must be greater than or equal to 0, and strictly less
//	than number of bit-table bytes * Byte.SIZE.
//	reader – The Fst.BytesReader to read. It must be positioned at the beginning of the bit-table.
func isBitSet(bitIndex int, reader BytesReader) (bool, error) {
	//if err := assert(bitIndex >= 0,
	//	fmt.Sprintf("bitIndex=%d", bitIndex)); err != nil {
	//	return false, err
	//}

	if err := reader.SkipBytes(bitIndex >> 3); err != nil {
		return false, err
	}

	num, err := readByte(reader)
	if err != nil {
		return false, err
	}
	return num&(1<<(bitIndex&(BYTE_SIZE-1))) != 0, nil
}

// Counts all bits set in the bit-table.
// Params:
//
//	bitTableBytes – The number of bytes in the bit-table.
//	reader – The Fst.BytesReader to read. It must be positioned at the beginning of the bit-table.
func countBits(bitTableBytes int64, reader BytesReader) (int64, error) {
	//if err := assert(bitTableBytes >= 0,
	//	fmt.Sprintf("bitTableBytes=%d", bitTableBytes)); err != nil {
	//	return 0, err
	//}

	bitCount := 0
	for i := bitTableBytes >> 3; i > 0; i-- {
		// Count the bits set for all plain longs.
		num, err := read8Bytes(reader)
		if err != nil {
			return 0, err
		}
		bitCount += bits.OnesCount64(uint64(num))
	}

	numRemainingBytes := bitTableBytes & (LONG_BYTES - 1)
	if numRemainingBytes != 0 {
		values, err := readUpTo8Bytes(numRemainingBytes, reader)
		if err != nil {
			return 0, err
		}

		bitCount += bits.OnesCount64(uint64(values))
	}
	return int64(bitCount), nil
}

// Counts the bits set up to the given bit zero-based index, exclusive.
// In other words, how many 1s there are up to the bit at the given index excluded.
// Example: bitIndex 10 means the third bit on the right of the second byte.
// Params:
//
//	bitIndex – The bit zero-based index, exclusive. It must be greater than or equal to 0, and less
//	than or equal to number of bit-table bytes * Byte.SIZE.
//	reader – The Fst.BytesReader to read. It must be positioned at the beginning of the bit-table.
func countBitsUpTo(bitIndex int, reader BytesReader) (int, error) {
	//if err := assert(bitIndex > 0,
	//	fmt.Sprintf("bitIndex=%d", bitIndex)); err != nil {
	//	return 0, err
	//}

	bitCount := 0

	// 计算有多少组uint64
	// bitIndex >> 6 相当于 bitIndex / 8
	for i := bitIndex >> 6; i > 0; i-- {
		// Count the bits set for all plain longs.
		num, err := read8Bytes(reader)
		if err != nil {
			return 0, err
		}
		bitCount += bits.OnesCount64(uint64(num))
	}

	// 处理 bitIndex % 8 剩余的字节
	remainingBits := bitIndex & (LONG_BYTES - 1)
	if remainingBits != 0 {
		numRemainingBytes := (remainingBits + (BYTE_SIZE - 1)) >> 3
		// Prepare a mask with 1s on the right up to bitIndex exclusive.
		mask := int64(1<<bitIndex) - 1 // Shifts are mod 64.
		// Count the bits set only within the mask part, so up to bitIndex exclusive.

		num, err := readUpTo8Bytes(int64(numRemainingBytes), reader)
		if err != nil {
			return 0, err
		}

		bitCount += bits.OnesCount64(uint64(num & mask))
	}
	return bitCount, nil
}

// Returns the index of the next bit set following the given bit zero-based index.
// For example with bits 100011: the next bit set after index=-1 is at index=0;
// the next bit set after index=0 is at index=1; the next bit set after index=1 is at index=5;
// there is no next bit set after index=5.
// Params:
//
//	bitIndex – The bit zero-based index. It must be greater than or equal to -1, and strictly less
//	than number of bit-table bytes * Byte.SIZE. bitTableBytes – The number of bytes in the bit-table.
//	reader – The Fst.BytesReader to read. It must be positioned at the beginning of the bit-table.
//
// Returns:
//
//	The zero-based index of the next bit set after the provided bitIndex; or -1 if none.
func nextBitSet(bitIndex, bitTableBytes int, reader BytesReader) (int, error) {
	//if err := assert(bitIndex >= -1 && bitIndex < bitTableBytes*BYTE_SIZE,
	//	fmt.Sprintf("bitIndex=%d bitTableBytes=%d", bitIndex, bitTableBytes)); err != nil {
	//	return 0, err
	//}

	byteIndex := bitIndex / BYTE_SIZE
	mask := int32(-1) << ((bitIndex + 1) & (BYTE_SIZE - 1))
	i := int32(0)

	if mask == -1 && bitIndex != -1 {
		if err := reader.SkipBytes(byteIndex + 1); err != nil {
			return 0, err
		}
		i = 0
	} else {
		if err := reader.SkipBytes(byteIndex); err != nil {
			return 0, err
		}
		b, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}

		i = int32(b) & mask
	}
	for i == 0 {
		if byteIndex+1 == bitTableBytes {
			byteIndex++
			return -1, nil
		}
		b, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}
		i = int32(b)
	}

	return bits.TrailingZeros32(uint32(i)) + byteIndex<<3, nil
}

// Returns the index of the previous bit set preceding the given bit zero-based index.
// For example with bits 100011: there is no previous bit set before index=0.
// the previous bit set before index=1 is at index=0;
// the previous bit set before index=5 is at index=1; the previous bit set before index=64 is at index=5;
// Params:
//
//	bitIndex – The bit zero-based index. It must be greater than or equal to 0, and less
//	than or equal to number of bit-table bytes * Byte.SIZE.
//	reader – The Fst.BytesReader to read. It must be positioned at the beginning of the bit-table.
//
// Returns:
//
//	The zero-based index of the previous bit set before the provided bitIndex; or -1 if none.
func previousBitSet(bitIndex int, reader BytesReader) (int, error) {
	//if err := assert(bitIndex >= 0); err != nil {
	//	return 0, err
	//}

	byteIndex := bitIndex >> 3
	if err := reader.SkipBytes(byteIndex); err != nil {
		return 0, err
	}

	mask := uint32(1<<(bitIndex&(BYTE_SIZE-1))) - 1

	b, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}

	i := uint32(b) & mask
	for i == 0 {
		if byteIndex-1 == 0 {
			byteIndex--
			return -1, nil
		}

		// Fst.BytesReader implementations support negative skip.
		if err := reader.SkipBytes(-2); err != nil {
			return 0, err
		}

		b, err = reader.ReadByte()
		if err != nil {
			return 0, err
		}

		i = uint32(b)
	}

	return (INTEGER_SIZE - 1) - bits.LeadingZeros32(i) + (byteIndex << 3), nil
}

func readByte(reader BytesReader) (int64, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}
	return int64(b), nil
}

func readUpTo8Bytes(numBytes int64, reader BytesReader) (int64, error) {
	//if err := assert(numBytes > 0 && numBytes <= 8, fmt.Sprintf("numBytes=%d", numBytes)); err != nil {
	//	return 0, err
	//}

	bs := make([]byte, numBytes)
	if _, err := reader.Read(bs); err != nil {
		return 0, err
	}

	num := int64(0)
	shift := 0

	for _, v := range bs {
		num |= int64(v) << shift
		shift += 8
	}
	return num, nil
}

func read8Bytes(reader BytesReader) (int64, error) {
	bs := make([]byte, 8)
	if _, err := reader.Read(bs); err != nil {
		return 0, err
	}

	return int64(bs[0]) |
		int64(bs[1])<<8 |
		int64(bs[2])<<16 |
		int64(bs[3])<<24 |
		int64(bs[4])<<32 |
		int64(bs[5])<<40 |
		int64(bs[6])<<48 |
		int64(bs[7])<<56, nil
}
