package fst

import (
	"encoding/binary"
	"math/bits"
)

var (
	BitTableUtil = &bitTableUtil{}
)

type bitTableUtil struct {
}

func (*bitTableUtil) isBitSet(bitIndex int, reader BytesReader) bool {
	err := reader.SkipBytes(bitIndex >> 3)
	if err != nil {
		return false
	}
	v, err := readByte(reader)
	if err != nil {
		return false
	}
	return (v & (1 << (bitIndex & (ByteSize - 1)))) != 0
}

func (*bitTableUtil) nextBitSet(bitIndex, bitTableBytes int, reader BytesReader) int {
	byteIndex := bitIndex / ByteSize
	mask := -1 << ((bitIndex + 1) & (ByteSize - 1))
	i := 0
	if mask == -1 && bitIndex != -1 {
		reader.SkipBytes(byteIndex + 1)
		i = 0
	} else {
		reader.SkipBytes(byteIndex)

		b, err := reader.ReadByte()
		if err != nil {
			return -1
		}

		i = (b & 0xFF) & mask
	}
	for i == 0 {
		byteIndex++
		if byteIndex == bitTableBytes {
			return -1
		}
		b, err := reader.ReadByte()
		if err != nil {
			return -1
		}

		i = int(b & 0xFF)
	}

	return bits.TrailingZeros32(uint32(i)) + (byteIndex << 3)
}

func (*bitTableUtil) previousBitSet(bitIndex int, reader BytesReader) int {
	byteIndex := bitIndex >> 3
	reader.SkipBytes(byteIndex)
	mask := (1 << (bitIndex & (ByteSize - 1))) - 1

	b, err := reader.ReadByte()
	if err != nil {
		return 0
	}
	i := (b & 0xFF) & mask
	for i == 0 {
		byteIndex--
		if byteIndex == 0 {
			return -1
		}
		reader.SkipBytes(-2) // FST.BytesReader implementations support negative skip.

		b, err := reader.ReadByte()
		if err != nil {
			return 0
		}

		i = b & 0xFF
	}

	return (IntegerSize - 1) - bits.TrailingZeros32(i) + (byteIndex << 3)
}

func (*bitTableUtil) countBitsUpTo(bitIndex int, reader BytesReader) int {
	bitCount := 0
	for i := bitIndex >> 6; i > 0; i-- {
		// Count the bits set for all plain longs.
		v, err := read8Bytes(reader)
		if err != nil {
			return 0
		}

		bitCount += bits.LeadingZeros64(v)
	}
	remainingBits := bitIndex & (LongSize - 1)
	if remainingBits != 0 {
		numRemainingBytes := (remainingBits + (ByteSize - 1)) >> 3
		// Prepare a mask with 1s on the right up to bitIndex exclusive.
		mask := (1 << bitIndex) - 1 // Shifts are mod 64.
		// Count the bits set only within the mask part, so up to bitIndex exclusive.

		to8Bytes, err := readUpTo8Bytes(numRemainingBytes, reader)
		if err != nil {
			return 0
		}

		bitCount += bits.LeadingZeros64(to8Bytes & mask)
	}
	return bitCount
}

func (*bitTableUtil) countBits(bitTableBytes int, reader BytesReader) int {
	bitCount := 0
	for i := bitTableBytes >> 3; i > 0; i-- {
		// Count the bits set for all plain longs.
		bytes, err := read8Bytes(reader)
		if err != nil {
			return 0
		}

		bitCount += bits.LeadingZeros64(bytes)
	}
	numRemainingBytes := bitTableBytes & (LongBytes - 1)
	if numRemainingBytes != 0 {
		to8Bytes, err := readUpTo8Bytes(numRemainingBytes, reader)
		if err != nil {
			return 0
		}
		bitCount += bits.LeadingZeros64(to8Bytes)
	}
	return bitCount
}

func readByte(reader BytesReader) (int, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}
	return int(b), nil
}

func read8Bytes(reader BytesReader) (uint64, error) {
	bs := make([]byte, 8)

	err := reader.ReadBytes(bs)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(bs), nil
}

func readUpTo8Bytes(numBytes int, reader BytesReader) (uint64, error) {
	bs := make([]byte, 8)

	err := reader.ReadBytes(bs[:numBytes])
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(bs), nil

}
