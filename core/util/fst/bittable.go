package fst

import (
	"context"
	"encoding/binary"
	"math/bits"
)

// IsBitSet See BitTableUtil.IsBitSet(int, Fst.BytesReader).
func IsBitSet(ctx context.Context, bitIndex int, arc *Arc, in BytesReader) (bool, error) {
	if err := in.SetPosition(arc.bitTableStart); err != nil {
		return false, err
	}
	return isBitSet(ctx, bitIndex, in)
}

// CountBits See BitTableUtil.countBits(int, Fst.BytesReader).
// The count of bit set is the number of arcs of a direct addressing node.
func CountBits(arc *Arc, in BytesReader) (int, error) {
	if err := in.SetPosition(arc.bitTableStart); err != nil {
		return 0, err
	}

	numPresenceBytes := int64(getNumPresenceBytes(arc.NumArcs()))
	return countBits(numPresenceBytes, in)
}

// CountBitsUpTo See BitTableUtil.countBitsUpTo(int, Fst.BytesReader).
func CountBitsUpTo(bitIndex int, arc *Arc, in BytesReader) (int, error) {
	if err := in.SetPosition(arc.bitTableStart); err != nil {
		return 0, err
	}
	return countBitsUpTo(bitIndex, in)
}

// NextBitSet See BitTableUtil.NextBitSet(int, int, Fst.BytesReader).
func NextBitSet(ctx context.Context, bitIndex int, arc *Arc, in BytesReader) (int, error) {
	if err := in.SetPosition(arc.bitTableStart); err != nil {
		return 0, err
	}

	numBytes := getNumPresenceBytes(arc.NumArcs())
	return nextBitSet(ctx, bitIndex, numBytes, in)
}

// PreviousBitSet See BitTableUtil.previousBitSet(int, Fst.BytesReader).
func PreviousBitSet(bitIndex int, arc *Arc, in BytesReader) (int, error) {
	if err := in.SetPosition(arc.bitTableStart); err != nil {
		return 0, err
	}
	return previousBitSet(bitIndex, in)
}

const (
	BYTE_SIZE    = 8
	INTEGER_SIZE = 32
	LONG_SIZE    = 64
	LONG_BYTES   = 8
)

// Returns whether the bit at given zero-based index is set. Example: bitIndex 10 means the third
// bit on the right of the second byte.
//
//	bitIndex: The bit zero-based index. It must be greater than or equal to 0, and strictly less than number of bit-table bytes * BYTE_SIZE.
//	reader: The Fst.BytesReader to read. It must be positioned at the beginning of the bit-table.
func isBitSet(ctx context.Context, bitIndex int, reader BytesReader) (bool, error) {
	if err := reader.SkipBytes(ctx, bitIndex>>3); err != nil {
		return false, err
	}

	num, err := readByte(reader)
	if err != nil {
		return false, err
	}
	return num&(1<<(bitIndex&(BYTE_SIZE-1))) != 0, nil
}

// Counts all bits set in the bit-table.
// bitTableBytes: The number of bytes in the bit-table.
// reader: The Fst.BytesReader to read. It must be positioned at the beginning of the bit-table.
func countBits(bitTableBytes int64, reader BytesReader) (int, error) {
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
	return bitCount, nil
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
func nextBitSet(ctx context.Context, bitIndex, bitTableBytes int, reader BytesReader) (int, error) {
	byteIndex := bitIndex / BYTE_SIZE
	mask := int32(-1) << ((bitIndex + 1) & (BYTE_SIZE - 1))
	i := int32(0)

	if mask == -1 && bitIndex != -1 {
		if err := reader.SkipBytes(ctx, byteIndex+1); err != nil {
			return 0, err
		}
		i = 0
	} else {
		if err := reader.SkipBytes(ctx, byteIndex); err != nil {
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
	byteIndex := bitIndex >> 3
	if err := reader.SkipBytes(nil, byteIndex); err != nil {
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
		if err := reader.SkipBytes(nil, -2); err != nil {
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
	return int64(binary.LittleEndian.Uint64(bs)), nil
}
