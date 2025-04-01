package lucene84

import (
	"context"
	"encoding/binary"

	"github.com/geange/lucene-go/core/store"
)

const (
	FOR_UTIL_BLOCK_SIZE = 128
)

// ForUtil
// Inspired from https://fulmicoton.com/posts/bitpacking/
// Encodes multiple integers in a to get SIMD-like speedups.
// If bitsPerValue <= 8 then we pack 8 ints per long
// else if bitsPerValue <= 16 we pack 4 ints per long
// else we pack 2 ints per long
type ForUtil struct {
}

var (
	MASKS8  = make([]uint64, 8)
	MASKS16 = make([]uint64, 16)
	MASKS32 = make([]uint64, 32)

	MASK8_1   = uint64(0)
	MASK8_2   = uint64(0)
	MASK8_3   = uint64(0)
	MASK8_4   = uint64(0)
	MASK8_5   = uint64(0)
	MASK8_6   = uint64(0)
	MASK8_7   = uint64(0)
	MASK16_1  = uint64(0)
	MASK16_2  = uint64(0)
	MASK16_3  = uint64(0)
	MASK16_4  = uint64(0)
	MASK16_5  = uint64(0)
	MASK16_6  = uint64(0)
	MASK16_7  = uint64(0)
	MASK16_9  = uint64(0)
	MASK16_10 = uint64(0)
	MASK16_11 = uint64(0)
	MASK16_12 = uint64(0)
	MASK16_13 = uint64(0)
	MASK16_14 = uint64(0)
	MASK16_15 = uint64(0)
	MASK32_1  = uint64(0)
	MASK32_2  = uint64(0)
	MASK32_3  = uint64(0)
	MASK32_4  = uint64(0)
	MASK32_5  = uint64(0)
	MASK32_6  = uint64(0)
	MASK32_7  = uint64(0)
	MASK32_8  = uint64(0)
	MASK32_9  = uint64(0)
	MASK32_10 = uint64(0)
	MASK32_11 = uint64(0)
	MASK32_12 = uint64(0)
	MASK32_13 = uint64(0)
	MASK32_14 = uint64(0)
	MASK32_15 = uint64(0)
	MASK32_17 = uint64(0)
	MASK32_18 = uint64(0)
	MASK32_19 = uint64(0)
	MASK32_20 = uint64(0)
	MASK32_21 = uint64(0)
	MASK32_22 = uint64(0)
	MASK32_23 = uint64(0)
	MASK32_24 = uint64(0)
)

func maskInit() {
	for i := 0; i < 8; i++ {
		MASKS8[i] = mask8(i)
	}
	for i := 0; i < 16; i++ {
		MASKS16[i] = mask16(i)
	}
	for i := 0; i < 32; i++ {
		MASKS32[i] = mask32(i)
	}

	MASK8_1 = MASKS8[1]
	MASK8_2 = MASKS8[2]
	MASK8_3 = MASKS8[3]
	MASK8_4 = MASKS8[4]
	MASK8_5 = MASKS8[5]
	MASK8_6 = MASKS8[6]
	MASK8_7 = MASKS8[7]
	MASK16_1 = MASKS16[1]
	MASK16_2 = MASKS16[2]
	MASK16_3 = MASKS16[3]
	MASK16_4 = MASKS16[4]
	MASK16_5 = MASKS16[5]
	MASK16_6 = MASKS16[6]
	MASK16_7 = MASKS16[7]
	MASK16_9 = MASKS16[9]
	MASK16_10 = MASKS16[10]
	MASK16_11 = MASKS16[11]
	MASK16_12 = MASKS16[12]
	MASK16_13 = MASKS16[13]
	MASK16_14 = MASKS16[14]
	MASK16_15 = MASKS16[15]
	MASK32_1 = MASKS32[1]
	MASK32_2 = MASKS32[2]
	MASK32_3 = MASKS32[3]
	MASK32_4 = MASKS32[4]
	MASK32_5 = MASKS32[5]
	MASK32_6 = MASKS32[6]
	MASK32_7 = MASKS32[7]
	MASK32_8 = MASKS32[8]
	MASK32_9 = MASKS32[9]
	MASK32_10 = MASKS32[10]
	MASK32_11 = MASKS32[11]
	MASK32_12 = MASKS32[12]
	MASK32_13 = MASKS32[13]
	MASK32_14 = MASKS32[14]
	MASK32_15 = MASKS32[15]
	MASK32_17 = MASKS32[17]
	MASK32_18 = MASKS32[18]
	MASK32_19 = MASKS32[19]
	MASK32_20 = MASKS32[20]
	MASK32_21 = MASKS32[21]
	MASK32_22 = MASKS32[22]
	MASK32_23 = MASKS32[23]
	MASK32_24 = MASKS32[24]
}

func expandMask32(mask32 uint64) uint64 {
	return mask32 | (mask32 << 32)
}

func expandMask16(mask16 uint64) uint64 {
	return expandMask32(mask16 | (mask16 << 16))
}

func expandMask8(mask8 uint64) uint64 {
	return expandMask16(mask8 | (mask8 << 8))
}

func mask8(bitsPerValue int) uint64 {
	return expandMask8(uint64(1<<bitsPerValue) - 1)
}

func mask16(bitsPerValue int) uint64 {
	return expandMask16(uint64(1<<bitsPerValue) - 1)
}

func mask32(bitsPerValue int) uint64 {
	return expandMask32(uint64(1<<bitsPerValue) - 1)
}

func expand8(arr []uint64) {
	for i := 0; i < 16; i++ {
		l := arr[i]
		arr[i] = (l >> 56) & 0xFF
		arr[16+i] = (l >> 48) & 0xFF
		arr[32+i] = (l >> 40) & 0xFF
		arr[48+i] = (l >> 32) & 0xFF
		arr[64+i] = (l >> 24) & 0xFF
		arr[80+i] = (l >> 16) & 0xFF
		arr[96+i] = (l >> 8) & 0xFF
		arr[112+i] = l & 0xFF
	}
}

func expand8To32(arr []uint64) {
	for i := 0; i < 16; i++ {
		l := arr[i]
		arr[i] = (l >> 24) & 0x000000FF000000FF
		arr[16+i] = (l >> 16) & 0x000000FF000000FF
		arr[32+i] = (l >> 8) & 0x000000FF000000FF
		arr[48+i] = l & 0x000000FF000000FF
	}
}

func collapse8(arr []uint64) {
	for i := 0; i < 16; i++ {
		arr[i] = (arr[i] << 56) |
			(arr[16+i] << 48) |
			(arr[32+i] << 40) |
			(arr[48+i] << 32) |
			(arr[64+i] << 24) |
			(arr[80+i] << 16) |
			(arr[96+i] << 8) |
			arr[112+i]
	}
}

func expand16(arr []uint64) {
	for i := 0; i < 32; i++ {
		l := arr[i]
		arr[i] = (l >> 48) & 0xFFFF
		arr[32+i] = (l >> 32) & 0xFFFF
		arr[64+i] = (l >> 16) & 0xFFFF
		arr[96+i] = l & 0xFFFF
	}
}

func expand16To32(arr []uint64) {
	for i := 0; i < 32; i++ {
		l := arr[i]
		arr[i] = (l >> 16) & 0x0000FFFF0000FFFF
		arr[32+i] = l & 0x0000FFFF0000FFFF
	}
}

func shiftLongs(src []uint64, count int, dest []uint64, index, shift int, mask uint64) {
	for i := 0; i < count; i++ {
		dest[index+i] = (src[i] >> shift) & mask
	}
}

func collapse16(arr []uint64) {
	for i := 0; i < 32; i++ {
		arr[i] = (arr[i] << 48) |
			(arr[32+i] << 32) |
			(arr[64+i] << 16) |
			arr[96+i]
	}
}

func expand32(arr []uint64) {
	for i := 0; i < 64; i++ {
		l := arr[i]
		arr[i] = l >> 32
		arr[64+i] = l & 0xFFFFFFFF
	}
}

func collapse32(arr []uint64) {
	for i := 0; i < 64; i++ {
		arr[i] = (arr[i] << 32) | arr[64+i]
	}
}

type IntCodec struct {
	tmp []uint64
}

// Encode 128 integers from longs into out.
func (e *IntCodec) Encode(longs []uint64, bitsPerValue int, out store.DataOutput) error {
	var nextPrimitive int
	var numLongs int
	if bitsPerValue <= 8 {
		nextPrimitive = 8
		numLongs = BLOCK_SIZE / 8
		collapse8(longs)
	} else if bitsPerValue <= 16 {
		nextPrimitive = 16
		numLongs = BLOCK_SIZE / 4
		collapse16(longs)
	} else {
		nextPrimitive = 32
		numLongs = BLOCK_SIZE / 2
		collapse32(longs)
	}

	numLongsPerShift := bitsPerValue * 2
	idx := 0

	shift := nextPrimitive - bitsPerValue
	for i := 0; i < numLongsPerShift; i++ {
		e.tmp[i] = longs[idx] << shift
		idx++
	}
	for shift = shift - bitsPerValue; shift >= 0; shift -= bitsPerValue {
		for i := 0; i < numLongsPerShift; i++ {
			e.tmp[i] |= longs[idx] << shift
			idx++
		}
	}

	remainingBitsPerLong := shift + bitsPerValue
	var maskRemainingBitsPerLong uint64
	if nextPrimitive == 8 {
		maskRemainingBitsPerLong = MASKS8[remainingBitsPerLong]
	} else if nextPrimitive == 16 {
		maskRemainingBitsPerLong = MASKS16[remainingBitsPerLong]
	} else {
		maskRemainingBitsPerLong = MASKS32[remainingBitsPerLong]
	}

	tmpIdx := 0
	remainingBitsPerValue := bitsPerValue

	for idx < numLongs {
		if remainingBitsPerValue >= remainingBitsPerLong {
			remainingBitsPerValue -= remainingBitsPerLong
			e.tmp[tmpIdx] |= (longs[idx] >> remainingBitsPerValue) & maskRemainingBitsPerLong
			tmpIdx++
			if remainingBitsPerValue == 0 {
				idx++
				remainingBitsPerValue = bitsPerValue
			}
			continue
		}

		var mask1, mask2 uint64
		if nextPrimitive == 8 {
			mask1 = MASKS8[remainingBitsPerValue]
			mask2 = MASKS8[remainingBitsPerLong-remainingBitsPerValue]
		} else if nextPrimitive == 16 {
			mask1 = MASKS16[remainingBitsPerValue]
			mask2 = MASKS16[remainingBitsPerLong-remainingBitsPerValue]
		} else {
			mask1 = MASKS32[remainingBitsPerValue]
			mask2 = MASKS32[remainingBitsPerLong-remainingBitsPerValue]
		}
		e.tmp[tmpIdx] |= (longs[idx] & mask1) << (remainingBitsPerLong - remainingBitsPerValue)
		idx++
		remainingBitsPerValue = bitsPerValue - remainingBitsPerLong + remainingBitsPerValue
		e.tmp[tmpIdx] |= (longs[idx] >> remainingBitsPerValue) & mask2
		tmpIdx++
	}

	bs := make([]byte, 8)
	for i := 0; i < numLongsPerShift; i++ {
		// Java longs are big endian and we want to read little endian longs, so we need to reverse bytes
		binary.LittleEndian.PutUint64(bs, e.tmp[i])
		if _, err := out.Write(bs); err != nil {
			return err
		}
	}
	return nil
}

func (e *IntCodec) decode1(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[:2]); err != nil {
		return err
	}
	shiftLongs(tmp, 2, longs, 0, 7, MASK8_1)
	shiftLongs(tmp, 2, longs, 2, 6, MASK8_1)
	shiftLongs(tmp, 2, longs, 4, 5, MASK8_1)
	shiftLongs(tmp, 2, longs, 6, 4, MASK8_1)
	shiftLongs(tmp, 2, longs, 8, 3, MASK8_1)
	shiftLongs(tmp, 2, longs, 10, 2, MASK8_1)
	shiftLongs(tmp, 2, longs, 12, 1, MASK8_1)
	shiftLongs(tmp, 2, longs, 14, 0, MASK8_1)
	return nil
}

func (e *IntCodec) decode2(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[:4]); err != nil {
		return err
	}
	shiftLongs(tmp, 4, longs, 0, 6, MASK8_2)
	shiftLongs(tmp, 4, longs, 4, 4, MASK8_2)
	shiftLongs(tmp, 4, longs, 8, 2, MASK8_2)
	shiftLongs(tmp, 4, longs, 12, 0, MASK8_2)
	return nil
}

func (e *IntCodec) decode3(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:6]); err != nil {
		return err
	}
	shiftLongs(tmp, 6, longs, 0, 5, MASK8_3)
	shiftLongs(tmp, 6, longs, 6, 2, MASK8_3)
	iter := 0
	tmpIdx := 0
	longsIdx := 12
	for iter < 2 {
		l0 := (tmp[tmpIdx+0] & MASK8_2) << 1
		l0 |= (tmp[tmpIdx+1] >> 1) & MASK8_1
		longs[longsIdx+0] = l0
		l1 := (tmp[tmpIdx+1] & MASK8_1) << 2
		l1 |= (tmp[tmpIdx+2] & MASK8_2) << 0
		longs[longsIdx+1] = l1

		iter++
		tmpIdx += 3
		longsIdx += 2
	}
	return nil
}

func (e *IntCodec) decode4(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:8]); err != nil {
		return err
	}
	shiftLongs(tmp, 8, longs, 0, 4, MASK8_4)
	shiftLongs(tmp, 8, longs, 8, 0, MASK8_4)
	return nil
}

func (e *IntCodec) decode5(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:10]); err != nil {
		return err
	}
	shiftLongs(tmp, 10, longs, 0, 3, MASK8_5)
	iter := 0
	tmpIdx := 0
	longsIdx := 10
	for iter < 2 {
		l0 := (tmp[tmpIdx+0] & MASK8_3) << 2
		l0 |= (tmp[tmpIdx+1] >> 1) & MASK8_2
		longs[longsIdx+0] = l0
		l1 := (tmp[tmpIdx+1] & MASK8_1) << 4
		l1 |= (tmp[tmpIdx+2] & MASK8_3) << 1
		l1 |= (tmp[tmpIdx+3] >> 2) & MASK8_1
		longs[longsIdx+1] = l1
		l2 := (tmp[tmpIdx+3] & MASK8_2) << 3
		l2 |= (tmp[tmpIdx+4] & MASK8_3) << 0
		longs[longsIdx+2] = l2

		iter++
		tmpIdx += 5
		longsIdx += 3
	}
	return nil
}

func (e *IntCodec) decode6(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:12]); err != nil {
		return err
	}
	shiftLongs(tmp, 12, longs, 0, 2, MASK8_6)
	shiftLongs(tmp, 12, tmp, 0, 0, MASK8_2)
	iter := 0
	tmpIdx := 0
	longsIdx := 12
	for iter < 4 {
		l0 := tmp[tmpIdx+0] << 4
		l0 |= tmp[tmpIdx+1] << 2
		l0 |= tmp[tmpIdx+2] << 0
		longs[longsIdx+0] = l0

		iter++
		tmpIdx += 3
		longsIdx += 1
	}
	return nil
}

func (e *IntCodec) decode7(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:14]); err != nil {
		return err
	}
	shiftLongs(tmp, 14, longs, 0, 1, MASK8_7)
	shiftLongs(tmp, 14, tmp, 0, 0, MASK8_1)
	iter := 0
	tmpIdx := 0
	longsIdx := 14
	for iter < 2 {
		l0 := tmp[tmpIdx+0] << 6
		l0 |= tmp[tmpIdx+1] << 5
		l0 |= tmp[tmpIdx+2] << 4
		l0 |= tmp[tmpIdx+3] << 3
		l0 |= tmp[tmpIdx+4] << 2
		l0 |= tmp[tmpIdx+5] << 1
		l0 |= tmp[tmpIdx+6] << 0
		longs[longsIdx+0] = l0

		iter++
		tmpIdx += 7
		longsIdx += 1
	}
	return nil
}

func (e *IntCodec) decode8(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	return in.ReadLELongs(ctx, longs[0:16])
}

func (e *IntCodec) decode9(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[:18]); err != nil {
		return err
	}
	shiftLongs(tmp, 18, longs, 0, 7, MASK16_9)
	iter := 0
	tmpIdx := 0
	longsIdx := 18
	for iter < 2 {
		l0 := (tmp[tmpIdx+0] & MASK16_7) << 2
		l0 |= (tmp[tmpIdx+1] >> 5) & MASK16_2
		longs[longsIdx+0] = l0
		l1 := (tmp[tmpIdx+1] & MASK16_5) << 4
		l1 |= (tmp[tmpIdx+2] >> 3) & MASK16_4
		longs[longsIdx+1] = l1
		l2 := (tmp[tmpIdx+2] & MASK16_3) << 6
		l2 |= (tmp[tmpIdx+3] >> 1) & MASK16_6
		longs[longsIdx+2] = l2
		l3 := (tmp[tmpIdx+3] & MASK16_1) << 8
		l3 |= (tmp[tmpIdx+4] & MASK16_7) << 1
		l3 |= (tmp[tmpIdx+5] >> 6) & MASK16_1
		longs[longsIdx+3] = l3
		l4 := (tmp[tmpIdx+5] & MASK16_6) << 3
		l4 |= (tmp[tmpIdx+6] >> 4) & MASK16_3
		longs[longsIdx+4] = l4
		l5 := (tmp[tmpIdx+6] & MASK16_4) << 5
		l5 |= (tmp[tmpIdx+7] >> 2) & MASK16_5
		longs[longsIdx+5] = l5
		l6 := (tmp[tmpIdx+7] & MASK16_2) << 7
		l6 |= (tmp[tmpIdx+8] & MASK16_7) << 0
		longs[longsIdx+6] = l6

		iter++
		tmpIdx += 9
		longsIdx += 7
	}
	return nil
}

func (e *IntCodec) decode10(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[:20]); err != nil {
		return err
	}
	shiftLongs(tmp, 20, longs, 0, 6, MASK16_10)
	iter := 0
	tmpIdx := 0
	longsIdx := 20
	for iter < 4 {
		l0 := (tmp[tmpIdx+0] & MASK16_6) << 4
		l0 |= (tmp[tmpIdx+1] >> 2) & MASK16_4
		longs[longsIdx+0] = l0
		l1 := (tmp[tmpIdx+1] & MASK16_2) << 8
		l1 |= (tmp[tmpIdx+2] & MASK16_6) << 2
		l1 |= (tmp[tmpIdx+3] >> 4) & MASK16_2
		longs[longsIdx+1] = l1
		l2 := (tmp[tmpIdx+3] & MASK16_4) << 6
		l2 |= (tmp[tmpIdx+4] & MASK16_6) << 0
		longs[longsIdx+2] = l2

		iter++
		tmpIdx += 5
		longsIdx += 3
	}
	return nil
}

func (e *IntCodec) decode11(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[:22]); err != nil {
		return err
	}
	shiftLongs(tmp, 22, longs, 0, 5, MASK16_11)
	iter := 0
	tmpIdx := 0
	longsIdx := 22
	for iter < 2 {
		l0 := (tmp[tmpIdx+0] & MASK16_5) << 6
		l0 |= (tmp[tmpIdx+1] & MASK16_5) << 1
		l0 |= (tmp[tmpIdx+2] >> 4) & MASK16_1
		longs[longsIdx+0] = l0
		l1 := (tmp[tmpIdx+2] & MASK16_4) << 7
		l1 |= (tmp[tmpIdx+3] & MASK16_5) << 2
		l1 |= (tmp[tmpIdx+4] >> 3) & MASK16_2
		longs[longsIdx+1] = l1
		l2 := (tmp[tmpIdx+4] & MASK16_3) << 8
		l2 |= (tmp[tmpIdx+5] & MASK16_5) << 3
		l2 |= (tmp[tmpIdx+6] >> 2) & MASK16_3
		longs[longsIdx+2] = l2
		l3 := (tmp[tmpIdx+6] & MASK16_2) << 9
		l3 |= (tmp[tmpIdx+7] & MASK16_5) << 4
		l3 |= (tmp[tmpIdx+8] >> 1) & MASK16_4
		longs[longsIdx+3] = l3
		l4 := (tmp[tmpIdx+8] & MASK16_1) << 10
		l4 |= (tmp[tmpIdx+9] & MASK16_5) << 5
		l4 |= (tmp[tmpIdx+10] & MASK16_5) << 0
		longs[longsIdx+4] = l4

		iter++
		tmpIdx += 11
		longsIdx += 5
	}
	return nil
}

func (e *IntCodec) decode12(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:24]); err != nil {
		return err
	}
	shiftLongs(tmp, 24, longs, 0, 4, MASK16_12)
	shiftLongs(tmp, 24, tmp, 0, 0, MASK16_4)
	iter := 0
	tmpIdx := 0
	longsIdx := 24
	for iter < 8 {
		l0 := tmp[tmpIdx+0] << 8
		l0 |= tmp[tmpIdx+1] << 4
		l0 |= tmp[tmpIdx+2] << 0
		longs[longsIdx+0] = l0

		iter++
		tmpIdx += 3
		longsIdx += 1
	}
	return nil
}

func (e *IntCodec) decode13(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:26]); err != nil {
		return err
	}
	shiftLongs(tmp, 26, longs, 0, 3, MASK16_13)
	iter := 0
	tmpIdx := 0
	longsIdx := 26
	for iter < 2 {
		l0 := (tmp[tmpIdx+0] & MASK16_3) << 10
		l0 |= (tmp[tmpIdx+1] & MASK16_3) << 7
		l0 |= (tmp[tmpIdx+2] & MASK16_3) << 4
		l0 |= (tmp[tmpIdx+3] & MASK16_3) << 1
		l0 |= (tmp[tmpIdx+4] >> 2) & MASK16_1
		longs[longsIdx+0] = l0
		l1 := (tmp[tmpIdx+4] & MASK16_2) << 11
		l1 |= (tmp[tmpIdx+5] & MASK16_3) << 8
		l1 |= (tmp[tmpIdx+6] & MASK16_3) << 5
		l1 |= (tmp[tmpIdx+7] & MASK16_3) << 2
		l1 |= (tmp[tmpIdx+8] >> 1) & MASK16_2
		longs[longsIdx+1] = l1
		l2 := (tmp[tmpIdx+8] & MASK16_1) << 12
		l2 |= (tmp[tmpIdx+9] & MASK16_3) << 9
		l2 |= (tmp[tmpIdx+10] & MASK16_3) << 6
		l2 |= (tmp[tmpIdx+11] & MASK16_3) << 3
		l2 |= (tmp[tmpIdx+12] & MASK16_3) << 0
		longs[longsIdx+2] = l2

		iter++
		tmpIdx += 13
		longsIdx += 3
	}
	return nil
}

func (e *IntCodec) decode14(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:28]); err != nil {
		return err
	}
	shiftLongs(tmp, 28, longs, 0, 2, MASK16_14)
	shiftLongs(tmp, 28, tmp, 0, 0, MASK16_2)
	iter := 0
	tmpIdx := 0
	longsIdx := 28
	for iter < 4 {
		l0 := tmp[tmpIdx+0] << 12
		l0 |= tmp[tmpIdx+1] << 10
		l0 |= tmp[tmpIdx+2] << 8
		l0 |= tmp[tmpIdx+3] << 6
		l0 |= tmp[tmpIdx+4] << 4
		l0 |= tmp[tmpIdx+5] << 2
		l0 |= tmp[tmpIdx+6] << 0
		longs[longsIdx+0] = l0

		iter++
		tmpIdx += 7
		longsIdx += 1
	}
	return nil
}

func (e *IntCodec) decode15(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:30]); err != nil {
		return err
	}
	shiftLongs(tmp, 30, longs, 0, 1, MASK16_15)
	shiftLongs(tmp, 30, tmp, 0, 0, MASK16_1)
	iter := 0
	tmpIdx := 0
	longsIdx := 30
	for iter < 2 {
		l0 := tmp[tmpIdx+0] << 14
		l0 |= tmp[tmpIdx+1] << 13
		l0 |= tmp[tmpIdx+2] << 12
		l0 |= tmp[tmpIdx+3] << 11
		l0 |= tmp[tmpIdx+4] << 10
		l0 |= tmp[tmpIdx+5] << 9
		l0 |= tmp[tmpIdx+6] << 8
		l0 |= tmp[tmpIdx+7] << 7
		l0 |= tmp[tmpIdx+8] << 6
		l0 |= tmp[tmpIdx+9] << 5
		l0 |= tmp[tmpIdx+10] << 4
		l0 |= tmp[tmpIdx+11] << 3
		l0 |= tmp[tmpIdx+12] << 2
		l0 |= tmp[tmpIdx+13] << 1
		l0 |= tmp[tmpIdx+14] << 0
		longs[longsIdx+0] = l0

		iter++
		tmpIdx += 15
		longsIdx += 1
	}
	return nil
}

func (e *IntCodec) decode16(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	return in.ReadLELongs(ctx, longs[0:32])
}

func (e *IntCodec) decode17(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:34]); err != nil {
		return err
	}
	shiftLongs(tmp, 34, longs, 0, 15, MASK32_17)
	iter := 0
	tmpIdx := 0
	longsIdx := 34
	for iter < 2 {
		l0 := (tmp[tmpIdx+0] & MASK32_15) << 2
		l0 |= (tmp[tmpIdx+1] >> 13) & MASK32_2
		longs[longsIdx+0] = l0
		l1 := (tmp[tmpIdx+1] & MASK32_13) << 4
		l1 |= (tmp[tmpIdx+2] >> 11) & MASK32_4
		longs[longsIdx+1] = l1
		l2 := (tmp[tmpIdx+2] & MASK32_11) << 6
		l2 |= (tmp[tmpIdx+3] >> 9) & MASK32_6
		longs[longsIdx+2] = l2
		l3 := (tmp[tmpIdx+3] & MASK32_9) << 8
		l3 |= (tmp[tmpIdx+4] >> 7) & MASK32_8
		longs[longsIdx+3] = l3
		l4 := (tmp[tmpIdx+4] & MASK32_7) << 10
		l4 |= (tmp[tmpIdx+5] >> 5) & MASK32_10
		longs[longsIdx+4] = l4
		l5 := (tmp[tmpIdx+5] & MASK32_5) << 12
		l5 |= (tmp[tmpIdx+6] >> 3) & MASK32_12
		longs[longsIdx+5] = l5
		l6 := (tmp[tmpIdx+6] & MASK32_3) << 14
		l6 |= (tmp[tmpIdx+7] >> 1) & MASK32_14
		longs[longsIdx+6] = l6
		l7 := (tmp[tmpIdx+7] & MASK32_1) << 16
		l7 |= (tmp[tmpIdx+8] & MASK32_15) << 1
		l7 |= (tmp[tmpIdx+9] >> 14) & MASK32_1
		longs[longsIdx+7] = l7
		l8 := (tmp[tmpIdx+9] & MASK32_14) << 3
		l8 |= (tmp[tmpIdx+10] >> 12) & MASK32_3
		longs[longsIdx+8] = l8
		l9 := (tmp[tmpIdx+10] & MASK32_12) << 5
		l9 |= (tmp[tmpIdx+11] >> 10) & MASK32_5
		longs[longsIdx+9] = l9
		l10 := (tmp[tmpIdx+11] & MASK32_10) << 7
		l10 |= (tmp[tmpIdx+12] >> 8) & MASK32_7
		longs[longsIdx+10] = l10
		l11 := (tmp[tmpIdx+12] & MASK32_8) << 9
		l11 |= (tmp[tmpIdx+13] >> 6) & MASK32_9
		longs[longsIdx+11] = l11
		l12 := (tmp[tmpIdx+13] & MASK32_6) << 11
		l12 |= (tmp[tmpIdx+14] >> 4) & MASK32_11
		longs[longsIdx+12] = l12
		l13 := (tmp[tmpIdx+14] & MASK32_4) << 13
		l13 |= (tmp[tmpIdx+15] >> 2) & MASK32_13
		longs[longsIdx+13] = l13
		l14 := (tmp[tmpIdx+15] & MASK32_2) << 15
		l14 |= (tmp[tmpIdx+16] & MASK32_15) << 0
		longs[longsIdx+14] = l14

		iter++
		tmpIdx += 17
		longsIdx += 15
	}
	return nil
}

func (e *IntCodec) decode18(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:36]); err != nil {
		return err
	}
	shiftLongs(tmp, 36, longs, 0, 14, MASK32_18)
	iter := 0
	tmpIdx := 0
	longsIdx := 36
	for iter < 4 {
		l0 := (tmp[tmpIdx+0] & MASK32_14) << 4
		l0 |= (tmp[tmpIdx+1] >> 10) & MASK32_4
		longs[longsIdx+0] = l0
		l1 := (tmp[tmpIdx+1] & MASK32_10) << 8
		l1 |= (tmp[tmpIdx+2] >> 6) & MASK32_8
		longs[longsIdx+1] = l1
		l2 := (tmp[tmpIdx+2] & MASK32_6) << 12
		l2 |= (tmp[tmpIdx+3] >> 2) & MASK32_12
		longs[longsIdx+2] = l2
		l3 := (tmp[tmpIdx+3] & MASK32_2) << 16
		l3 |= (tmp[tmpIdx+4] & MASK32_14) << 2
		l3 |= (tmp[tmpIdx+5] >> 12) & MASK32_2
		longs[longsIdx+3] = l3
		l4 := (tmp[tmpIdx+5] & MASK32_12) << 6
		l4 |= (tmp[tmpIdx+6] >> 8) & MASK32_6
		longs[longsIdx+4] = l4
		l5 := (tmp[tmpIdx+6] & MASK32_8) << 10
		l5 |= (tmp[tmpIdx+7] >> 4) & MASK32_10
		longs[longsIdx+5] = l5
		l6 := (tmp[tmpIdx+7] & MASK32_4) << 14
		l6 |= (tmp[tmpIdx+8] & MASK32_14) << 0
		longs[longsIdx+6] = l6

		iter++
		tmpIdx += 9
		longsIdx += 7
	}
	return nil
}

func (e *IntCodec) decode19(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:38]); err != nil {
		return err
	}
	shiftLongs(tmp, 38, longs, 0, 13, MASK32_19)
	iter := 0
	tmpIdx := 0
	longsIdx := 38
	for iter < 2 {
		l0 := (tmp[tmpIdx+0] & MASK32_13) << 6
		l0 |= (tmp[tmpIdx+1] >> 7) & MASK32_6
		longs[longsIdx+0] = l0
		l1 := (tmp[tmpIdx+1] & MASK32_7) << 12
		l1 |= (tmp[tmpIdx+2] >> 1) & MASK32_12
		longs[longsIdx+1] = l1
		l2 := (tmp[tmpIdx+2] & MASK32_1) << 18
		l2 |= (tmp[tmpIdx+3] & MASK32_13) << 5
		l2 |= (tmp[tmpIdx+4] >> 8) & MASK32_5
		longs[longsIdx+2] = l2
		l3 := (tmp[tmpIdx+4] & MASK32_8) << 11
		l3 |= (tmp[tmpIdx+5] >> 2) & MASK32_11
		longs[longsIdx+3] = l3
		l4 := (tmp[tmpIdx+5] & MASK32_2) << 17
		l4 |= (tmp[tmpIdx+6] & MASK32_13) << 4
		l4 |= (tmp[tmpIdx+7] >> 9) & MASK32_4
		longs[longsIdx+4] = l4
		l5 := (tmp[tmpIdx+7] & MASK32_9) << 10
		l5 |= (tmp[tmpIdx+8] >> 3) & MASK32_10
		longs[longsIdx+5] = l5
		l6 := (tmp[tmpIdx+8] & MASK32_3) << 16
		l6 |= (tmp[tmpIdx+9] & MASK32_13) << 3
		l6 |= (tmp[tmpIdx+10] >> 10) & MASK32_3
		longs[longsIdx+6] = l6
		l7 := (tmp[tmpIdx+10] & MASK32_10) << 9
		l7 |= (tmp[tmpIdx+11] >> 4) & MASK32_9
		longs[longsIdx+7] = l7
		l8 := (tmp[tmpIdx+11] & MASK32_4) << 15
		l8 |= (tmp[tmpIdx+12] & MASK32_13) << 2
		l8 |= (tmp[tmpIdx+13] >> 11) & MASK32_2
		longs[longsIdx+8] = l8
		l9 := (tmp[tmpIdx+13] & MASK32_11) << 8
		l9 |= (tmp[tmpIdx+14] >> 5) & MASK32_8
		longs[longsIdx+9] = l9
		l10 := (tmp[tmpIdx+14] & MASK32_5) << 14
		l10 |= (tmp[tmpIdx+15] & MASK32_13) << 1
		l10 |= (tmp[tmpIdx+16] >> 12) & MASK32_1
		longs[longsIdx+10] = l10
		l11 := (tmp[tmpIdx+16] & MASK32_12) << 7
		l11 |= (tmp[tmpIdx+17] >> 6) & MASK32_7
		longs[longsIdx+11] = l11
		l12 := (tmp[tmpIdx+17] & MASK32_6) << 13
		l12 |= (tmp[tmpIdx+18] & MASK32_13) << 0
		longs[longsIdx+12] = l12

		iter++
		tmpIdx += 19
		longsIdx += 13
	}
	return nil
}

func (e *IntCodec) decode20(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:40]); err != nil {
		return err
	}
	shiftLongs(tmp, 40, longs, 0, 12, MASK32_20)
	iter := 0
	tmpIdx := 0
	longsIdx := 40
	for iter < 8 {
		l0 := (tmp[tmpIdx+0] & MASK32_12) << 8
		l0 |= (tmp[tmpIdx+1] >> 4) & MASK32_8
		longs[longsIdx+0] = l0
		l1 := (tmp[tmpIdx+1] & MASK32_4) << 16
		l1 |= (tmp[tmpIdx+2] & MASK32_12) << 4
		l1 |= (tmp[tmpIdx+3] >> 8) & MASK32_4
		longs[longsIdx+1] = l1
		l2 := (tmp[tmpIdx+3] & MASK32_8) << 12
		l2 |= (tmp[tmpIdx+4] & MASK32_12) << 0
		longs[longsIdx+2] = l2

		iter++
		tmpIdx += 5
		longsIdx += 3
	}
	return nil
}

func (e *IntCodec) decode21(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:42]); err != nil {
		return err
	}
	shiftLongs(tmp, 42, longs, 0, 11, MASK32_21)
	iter := 0
	tmpIdx := 0
	longsIdx := 42
	for iter < 2 {
		l0 := (tmp[tmpIdx+0] & MASK32_11) << 10
		l0 |= (tmp[tmpIdx+1] >> 1) & MASK32_10
		longs[longsIdx+0] = l0
		l1 := (tmp[tmpIdx+1] & MASK32_1) << 20
		l1 |= (tmp[tmpIdx+2] & MASK32_11) << 9
		l1 |= (tmp[tmpIdx+3] >> 2) & MASK32_9
		longs[longsIdx+1] = l1
		l2 := (tmp[tmpIdx+3] & MASK32_2) << 19
		l2 |= (tmp[tmpIdx+4] & MASK32_11) << 8
		l2 |= (tmp[tmpIdx+5] >> 3) & MASK32_8
		longs[longsIdx+2] = l2
		l3 := (tmp[tmpIdx+5] & MASK32_3) << 18
		l3 |= (tmp[tmpIdx+6] & MASK32_11) << 7
		l3 |= (tmp[tmpIdx+7] >> 4) & MASK32_7
		longs[longsIdx+3] = l3
		l4 := (tmp[tmpIdx+7] & MASK32_4) << 17
		l4 |= (tmp[tmpIdx+8] & MASK32_11) << 6
		l4 |= (tmp[tmpIdx+9] >> 5) & MASK32_6
		longs[longsIdx+4] = l4
		l5 := (tmp[tmpIdx+9] & MASK32_5) << 16
		l5 |= (tmp[tmpIdx+10] & MASK32_11) << 5
		l5 |= (tmp[tmpIdx+11] >> 6) & MASK32_5
		longs[longsIdx+5] = l5
		l6 := (tmp[tmpIdx+11] & MASK32_6) << 15
		l6 |= (tmp[tmpIdx+12] & MASK32_11) << 4
		l6 |= (tmp[tmpIdx+13] >> 7) & MASK32_4
		longs[longsIdx+6] = l6
		l7 := (tmp[tmpIdx+13] & MASK32_7) << 14
		l7 |= (tmp[tmpIdx+14] & MASK32_11) << 3
		l7 |= (tmp[tmpIdx+15] >> 8) & MASK32_3
		longs[longsIdx+7] = l7
		l8 := (tmp[tmpIdx+15] & MASK32_8) << 13
		l8 |= (tmp[tmpIdx+16] & MASK32_11) << 2
		l8 |= (tmp[tmpIdx+17] >> 9) & MASK32_2
		longs[longsIdx+8] = l8
		l9 := (tmp[tmpIdx+17] & MASK32_9) << 12
		l9 |= (tmp[tmpIdx+18] & MASK32_11) << 1
		l9 |= (tmp[tmpIdx+19] >> 10) & MASK32_1
		longs[longsIdx+9] = l9
		l10 := (tmp[tmpIdx+19] & MASK32_10) << 11
		l10 |= (tmp[tmpIdx+20] & MASK32_11) << 0
		longs[longsIdx+10] = l10

		iter++
		tmpIdx += 21
		longsIdx += 11
	}
	return nil
}

func (e *IntCodec) decode22(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:44]); err != nil {
		return err
	}
	shiftLongs(tmp, 44, longs, 0, 10, MASK32_22)
	iter := 0
	tmpIdx := 0
	longsIdx := 44
	for iter < 4 {
		l0 := (tmp[tmpIdx+0] & MASK32_10) << 12
		l0 |= (tmp[tmpIdx+1] & MASK32_10) << 2
		l0 |= (tmp[tmpIdx+2] >> 8) & MASK32_2
		longs[longsIdx+0] = l0
		l1 := (tmp[tmpIdx+2] & MASK32_8) << 14
		l1 |= (tmp[tmpIdx+3] & MASK32_10) << 4
		l1 |= (tmp[tmpIdx+4] >> 6) & MASK32_4
		longs[longsIdx+1] = l1
		l2 := (tmp[tmpIdx+4] & MASK32_6) << 16
		l2 |= (tmp[tmpIdx+5] & MASK32_10) << 6
		l2 |= (tmp[tmpIdx+6] >> 4) & MASK32_6
		longs[longsIdx+2] = l2
		l3 := (tmp[tmpIdx+6] & MASK32_4) << 18
		l3 |= (tmp[tmpIdx+7] & MASK32_10) << 8
		l3 |= (tmp[tmpIdx+8] >> 2) & MASK32_8
		longs[longsIdx+3] = l3
		l4 := (tmp[tmpIdx+8] & MASK32_2) << 20
		l4 |= (tmp[tmpIdx+9] & MASK32_10) << 10
		l4 |= (tmp[tmpIdx+10] & MASK32_10) << 0
		longs[longsIdx+4] = l4

		iter++
		tmpIdx += 11
		longsIdx += 5
	}
	return nil
}

func (e *IntCodec) decode23(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:46]); err != nil {
		return err
	}
	shiftLongs(tmp, 46, longs, 0, 9, MASK32_23)
	iter := 0
	tmpIdx := 0
	longsIdx := 46
	for iter < 2 {
		l0 := (tmp[tmpIdx+0] & MASK32_9) << 14
		l0 |= (tmp[tmpIdx+1] & MASK32_9) << 5
		l0 |= (tmp[tmpIdx+2] >> 4) & MASK32_5
		longs[longsIdx+0] = l0
		l1 := (tmp[tmpIdx+2] & MASK32_4) << 19
		l1 |= (tmp[tmpIdx+3] & MASK32_9) << 10
		l1 |= (tmp[tmpIdx+4] & MASK32_9) << 1
		l1 |= (tmp[tmpIdx+5] >> 8) & MASK32_1
		longs[longsIdx+1] = l1
		l2 := (tmp[tmpIdx+5] & MASK32_8) << 15
		l2 |= (tmp[tmpIdx+6] & MASK32_9) << 6
		l2 |= (tmp[tmpIdx+7] >> 3) & MASK32_6
		longs[longsIdx+2] = l2
		l3 := (tmp[tmpIdx+7] & MASK32_3) << 20
		l3 |= (tmp[tmpIdx+8] & MASK32_9) << 11
		l3 |= (tmp[tmpIdx+9] & MASK32_9) << 2
		l3 |= (tmp[tmpIdx+10] >> 7) & MASK32_2
		longs[longsIdx+3] = l3
		l4 := (tmp[tmpIdx+10] & MASK32_7) << 16
		l4 |= (tmp[tmpIdx+11] & MASK32_9) << 7
		l4 |= (tmp[tmpIdx+12] >> 2) & MASK32_7
		longs[longsIdx+4] = l4
		l5 := (tmp[tmpIdx+12] & MASK32_2) << 21
		l5 |= (tmp[tmpIdx+13] & MASK32_9) << 12
		l5 |= (tmp[tmpIdx+14] & MASK32_9) << 3
		l5 |= (tmp[tmpIdx+15] >> 6) & MASK32_3
		longs[longsIdx+5] = l5
		l6 := (tmp[tmpIdx+15] & MASK32_6) << 17
		l6 |= (tmp[tmpIdx+16] & MASK32_9) << 8
		l6 |= (tmp[tmpIdx+17] >> 1) & MASK32_8
		longs[longsIdx+6] = l6
		l7 := (tmp[tmpIdx+17] & MASK32_1) << 22
		l7 |= (tmp[tmpIdx+18] & MASK32_9) << 13
		l7 |= (tmp[tmpIdx+19] & MASK32_9) << 4
		l7 |= (tmp[tmpIdx+20] >> 5) & MASK32_4
		longs[longsIdx+7] = l7
		l8 := (tmp[tmpIdx+20] & MASK32_5) << 18
		l8 |= (tmp[tmpIdx+21] & MASK32_9) << 9
		l8 |= (tmp[tmpIdx+22] & MASK32_9) << 0
		longs[longsIdx+8] = l8

		iter++
		tmpIdx += 23
		longsIdx += 9
	}
	return nil
}

func (e *IntCodec) decode24(ctx context.Context, in store.DataInput, tmp, longs []uint64) error {
	if err := in.ReadLELongs(ctx, tmp[0:48]); err != nil {
		return err
	}
	shiftLongs(tmp, 48, longs, 0, 8, MASK32_24)
	shiftLongs(tmp, 48, tmp, 0, 0, MASK32_8)
	iter := 0
	tmpIdx := 0
	longsIdx := 48
	for iter < 16 {
		l0 := tmp[tmpIdx+0] << 16
		l0 |= tmp[tmpIdx+1] << 8
		l0 |= tmp[tmpIdx+2] << 0
		longs[longsIdx+0] = l0

		iter++
		tmpIdx += 3
		longsIdx += 1
	}
	return nil
}
