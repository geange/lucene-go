package packed

type BulkOperationPacked21 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked21() *BulkOperationPacked21 {
	return &BulkOperationPacked21{NewBulkOperationPacked(21)}
}

func (b *BulkOperationPacked21) DecodeLongToLong(blocks, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 43
		valuesOffset++
		values[valuesOffset] = (block0 >> 22) & 2097151
		valuesOffset++
		values[valuesOffset] = (block0 >> 1) & 2097151
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 1) << 20) | (block1 >> 44)
		valuesOffset++
		values[valuesOffset] = (block1 >> 23) & 2097151
		valuesOffset++
		values[valuesOffset] = (block1 >> 2) & 2097151
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 19) | (block2 >> 45)
		valuesOffset++
		values[valuesOffset] = (block2 >> 24) & 2097151
		valuesOffset++
		values[valuesOffset] = (block2 >> 3) & 2097151
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 7) << 18) | (block3 >> 46)
		valuesOffset++
		values[valuesOffset] = (block3 >> 25) & 2097151
		valuesOffset++
		values[valuesOffset] = (block3 >> 4) & 2097151
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 15) << 17) | (block4 >> 47)
		valuesOffset++
		values[valuesOffset] = (block4 >> 26) & 2097151
		valuesOffset++
		values[valuesOffset] = (block4 >> 5) & 2097151
		valuesOffset++
		block5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block4 & 31) << 16) | (block5 >> 48)
		valuesOffset++
		values[valuesOffset] = (block5 >> 27) & 2097151
		valuesOffset++
		values[valuesOffset] = (block5 >> 6) & 2097151
		valuesOffset++
		block6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block5 & 63) << 15) | (block6 >> 49)
		valuesOffset++
		values[valuesOffset] = (block6 >> 28) & 2097151
		valuesOffset++
		values[valuesOffset] = (block6 >> 7) & 2097151
		valuesOffset++
		block7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block6 & 127) << 14) | (block7 >> 50)
		valuesOffset++
		values[valuesOffset] = (block7 >> 29) & 2097151
		valuesOffset++
		values[valuesOffset] = (block7 >> 8) & 2097151
		valuesOffset++
		block8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block7 & 255) << 13) | (block8 >> 51)
		valuesOffset++
		values[valuesOffset] = (block8 >> 30) & 2097151
		valuesOffset++
		values[valuesOffset] = (block8 >> 9) & 2097151
		valuesOffset++
		block9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block8 & 511) << 12) | (block9 >> 52)
		valuesOffset++
		values[valuesOffset] = (block9 >> 31) & 2097151
		valuesOffset++
		values[valuesOffset] = (block9 >> 10) & 2097151
		valuesOffset++
		block10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block9 & 1023) << 11) | (block10 >> 53)
		valuesOffset++
		values[valuesOffset] = (block10 >> 32) & 2097151
		valuesOffset++
		values[valuesOffset] = (block10 >> 11) & 2097151
		valuesOffset++
		block11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block10 & 2047) << 10) | (block11 >> 54)
		valuesOffset++
		values[valuesOffset] = (block11 >> 33) & 2097151
		valuesOffset++
		values[valuesOffset] = (block11 >> 12) & 2097151
		valuesOffset++
		block12 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block11 & 4095) << 9) | (block12 >> 55)
		valuesOffset++
		values[valuesOffset] = (block12 >> 34) & 2097151
		valuesOffset++
		values[valuesOffset] = (block12 >> 13) & 2097151
		valuesOffset++
		block13 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block12 & 8191) << 8) | (block13 >> 56)
		valuesOffset++
		values[valuesOffset] = (block13 >> 35) & 2097151
		valuesOffset++
		values[valuesOffset] = (block13 >> 14) & 2097151
		valuesOffset++
		block14 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block13 & 16383) << 7) | (block14 >> 57)
		valuesOffset++
		values[valuesOffset] = (block14 >> 36) & 2097151
		valuesOffset++
		values[valuesOffset] = (block14 >> 15) & 2097151
		valuesOffset++
		block15 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block14 & 32767) << 6) | (block15 >> 58)
		valuesOffset++
		values[valuesOffset] = (block15 >> 37) & 2097151
		valuesOffset++
		values[valuesOffset] = (block15 >> 16) & 2097151
		valuesOffset++
		block16 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block15 & 65535) << 5) | (block16 >> 59)
		valuesOffset++
		values[valuesOffset] = (block16 >> 38) & 2097151
		valuesOffset++
		values[valuesOffset] = (block16 >> 17) & 2097151
		valuesOffset++
		block17 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block16 & 131071) << 4) | (block17 >> 60)
		valuesOffset++
		values[valuesOffset] = (block17 >> 39) & 2097151
		valuesOffset++
		values[valuesOffset] = (block17 >> 18) & 2097151
		valuesOffset++
		block18 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block17 & 262143) << 3) | (block18 >> 61)
		valuesOffset++
		values[valuesOffset] = (block18 >> 40) & 2097151
		valuesOffset++
		values[valuesOffset] = (block18 >> 19) & 2097151
		valuesOffset++
		block19 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block18 & 524287) << 2) | (block19 >> 62)
		valuesOffset++
		values[valuesOffset] = (block19 >> 41) & 2097151
		valuesOffset++
		values[valuesOffset] = (block19 >> 20) & 2097151
		valuesOffset++
		block20 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block19 & 1048575) << 1) | (block20 >> 63)
		valuesOffset++
		values[valuesOffset] = (block20 >> 42) & 2097151
		valuesOffset++
		values[valuesOffset] = (block20 >> 21) & 2097151
		valuesOffset++
		values[valuesOffset] = block20 & 2097151
		valuesOffset++
	}
}

func (b *BulkOperationPacked21) DecodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte1 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte2 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = (byte0 << 13) | (byte1 << 5) | (byte2 >> 3)
		valuesOffset++
		byte3 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte4 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte5 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte2 & 7) << 18) | (byte3 << 10) | (byte4 << 2) | (byte5 >> 6)
		valuesOffset++
		byte6 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte7 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte5 & 63) << 15) | (byte6 << 7) | (byte7 >> 1)
		valuesOffset++
		byte8 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte9 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte10 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte7 & 1) << 20) | (byte8 << 12) | (byte9 << 4) | (byte10 >> 4)
		valuesOffset++
		byte11 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte12 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte13 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte10 & 15) << 17) | (byte11 << 9) | (byte12 << 1) | (byte13 >> 7)
		valuesOffset++
		byte14 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte15 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte13 & 127) << 14) | (byte14 << 6) | (byte15 >> 2)
		valuesOffset++
		byte16 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte17 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte18 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte15 & 3) << 19) | (byte16 << 11) | (byte17 << 3) | (byte18 >> 5)
		valuesOffset++
		byte19 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte20 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte18 & 31) << 16) | (byte19 << 8) | byte20
		valuesOffset++
	}
}

func (b *BulkOperationPacked21) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte1 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte2 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = (byte0 << 13) | (byte1 << 5) | (byte2 >> 3)
		valuesOffset++
		byte3 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte4 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte5 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte2 & 7) << 18) | (byte3 << 10) | (byte4 << 2) | (byte5 >> 6)
		valuesOffset++
		byte6 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte7 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte5 & 63) << 15) | (byte6 << 7) | (byte7 >> 1)
		valuesOffset++
		byte8 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte9 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte10 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte7 & 1) << 20) | (byte8 << 12) | (byte9 << 4) | (byte10 >> 4)
		valuesOffset++
		byte11 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte12 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte13 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte10 & 15) << 17) | (byte11 << 9) | (byte12 << 1) | (byte13 >> 7)
		valuesOffset++
		byte14 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte15 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte13 & 127) << 14) | (byte14 << 6) | (byte15 >> 2)
		valuesOffset++
		byte16 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte17 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte18 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte15 & 3) << 19) | (byte16 << 11) | (byte17 << 3) | (byte18 >> 5)
		valuesOffset++
		byte19 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte20 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte18 & 31) << 16) | (byte19 << 8) | byte20
		valuesOffset++
	}
}
