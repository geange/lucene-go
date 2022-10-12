package packed

type BulkOperationPacked19 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked19() *BulkOperationPacked19 {
	return &BulkOperationPacked19{NewBulkOperationPacked(19)}
}

func (b *BulkOperationPacked19) DecodeLongToLong(blocks, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 45
		valuesOffset++
		values[valuesOffset] = (block0 >> 26) & 524287
		valuesOffset++
		values[valuesOffset] = (block0 >> 7) & 524287
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 127) << 12) | (block1 >> 52)
		valuesOffset++
		values[valuesOffset] = (block1 >> 33) & 524287
		valuesOffset++
		values[valuesOffset] = (block1 >> 14) & 524287
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 16383) << 5) | (block2 >> 59)
		valuesOffset++
		values[valuesOffset] = (block2 >> 40) & 524287
		valuesOffset++
		values[valuesOffset] = (block2 >> 21) & 524287
		valuesOffset++
		values[valuesOffset] = (block2 >> 2) & 524287
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 3) << 17) | (block3 >> 47)
		valuesOffset++
		values[valuesOffset] = (block3 >> 28) & 524287
		valuesOffset++
		values[valuesOffset] = (block3 >> 9) & 524287
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 511) << 10) | (block4 >> 54)
		valuesOffset++
		values[valuesOffset] = (block4 >> 35) & 524287
		valuesOffset++
		values[valuesOffset] = (block4 >> 16) & 524287
		valuesOffset++
		block5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block4 & 65535) << 3) | (block5 >> 61)
		valuesOffset++
		values[valuesOffset] = (block5 >> 42) & 524287
		valuesOffset++
		values[valuesOffset] = (block5 >> 23) & 524287
		valuesOffset++
		values[valuesOffset] = (block5 >> 4) & 524287
		valuesOffset++
		block6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block5 & 15) << 15) | (block6 >> 49)
		valuesOffset++
		values[valuesOffset] = (block6 >> 30) & 524287
		valuesOffset++
		values[valuesOffset] = (block6 >> 11) & 524287
		valuesOffset++
		block7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block6 & 2047) << 8) | (block7 >> 56)
		valuesOffset++
		values[valuesOffset] = (block7 >> 37) & 524287
		valuesOffset++
		values[valuesOffset] = (block7 >> 18) & 524287
		valuesOffset++
		block8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block7 & 262143) << 1) | (block8 >> 63)
		valuesOffset++
		values[valuesOffset] = (block8 >> 44) & 524287
		valuesOffset++
		values[valuesOffset] = (block8 >> 25) & 524287
		valuesOffset++
		values[valuesOffset] = (block8 >> 6) & 524287
		valuesOffset++
		block9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block8 & 63) << 13) | (block9 >> 51)
		valuesOffset++
		values[valuesOffset] = (block9 >> 32) & 524287
		valuesOffset++
		values[valuesOffset] = (block9 >> 13) & 524287
		valuesOffset++
		block10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block9 & 8191) << 6) | (block10 >> 58)
		valuesOffset++
		values[valuesOffset] = (block10 >> 39) & 524287
		valuesOffset++
		values[valuesOffset] = (block10 >> 20) & 524287
		valuesOffset++
		values[valuesOffset] = (block10 >> 1) & 524287
		valuesOffset++
		block11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block10 & 1) << 18) | (block11 >> 46)
		valuesOffset++
		values[valuesOffset] = (block11 >> 27) & 524287
		valuesOffset++
		values[valuesOffset] = (block11 >> 8) & 524287
		valuesOffset++
		block12 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block11 & 255) << 11) | (block12 >> 53)
		valuesOffset++
		values[valuesOffset] = (block12 >> 34) & 524287
		valuesOffset++
		values[valuesOffset] = (block12 >> 15) & 524287
		valuesOffset++
		block13 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block12 & 32767) << 4) | (block13 >> 60)
		valuesOffset++
		values[valuesOffset] = (block13 >> 41) & 524287
		valuesOffset++
		values[valuesOffset] = (block13 >> 22) & 524287
		valuesOffset++
		values[valuesOffset] = (block13 >> 3) & 524287
		valuesOffset++
		block14 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block13 & 7) << 16) | (block14 >> 48)
		valuesOffset++
		values[valuesOffset] = (block14 >> 29) & 524287
		valuesOffset++
		values[valuesOffset] = (block14 >> 10) & 524287
		valuesOffset++
		block15 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block14 & 1023) << 9) | (block15 >> 55)
		valuesOffset++
		values[valuesOffset] = (block15 >> 36) & 524287
		valuesOffset++
		values[valuesOffset] = (block15 >> 17) & 524287
		valuesOffset++
		block16 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block15 & 131071) << 2) | (block16 >> 62)
		valuesOffset++
		values[valuesOffset] = (block16 >> 43) & 524287
		valuesOffset++
		values[valuesOffset] = (block16 >> 24) & 524287
		valuesOffset++
		values[valuesOffset] = (block16 >> 5) & 524287
		valuesOffset++
		block17 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block16 & 31) << 14) | (block17 >> 50)
		valuesOffset++
		values[valuesOffset] = (block17 >> 31) & 524287
		valuesOffset++
		values[valuesOffset] = (block17 >> 12) & 524287
		valuesOffset++
		block18 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block17 & 4095) << 7) | (block18 >> 57)
		valuesOffset++
		values[valuesOffset] = (block18 >> 38) & 524287
		valuesOffset++
		values[valuesOffset] = (block18 >> 19) & 524287
		valuesOffset++
		values[valuesOffset] = block18 & 524287
		valuesOffset++
	}
}

func (b *BulkOperationPacked19) DecodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := int64(blocks[blocksOffset]) & 0xFF
		byte1 := int64(blocks[blocksOffset]) & 0xFF
		byte2 := int64(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = (byte0 << 11) | (byte1 << 3) | (byte2 >> 5)
		valuesOffset++
		byte3 := int64(blocks[blocksOffset]) & 0xFF
		byte4 := int64(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte2 & 31) << 14) | (byte3 << 6) | (byte4 >> 2)
		valuesOffset++
		byte5 := int64(blocks[blocksOffset]) & 0xFF
		byte6 := int64(blocks[blocksOffset]) & 0xFF
		byte7 := int64(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte4 & 3) << 17) | (byte5 << 9) | (byte6 << 1) | (byte7 >> 7)
		valuesOffset++
		byte8 := int64(blocks[blocksOffset]) & 0xFF
		byte9 := int64(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte7 & 127) << 12) | (byte8 << 4) | (byte9 >> 4)
		valuesOffset++
		byte10 := int64(blocks[blocksOffset]) & 0xFF
		byte11 := int64(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte9 & 15) << 15) | (byte10 << 7) | (byte11 >> 1)
		valuesOffset++
		byte12 := int64(blocks[blocksOffset]) & 0xFF
		byte13 := int64(blocks[blocksOffset]) & 0xFF
		byte14 := int64(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte11 & 1) << 18) | (byte12 << 10) | (byte13 << 2) | (byte14 >> 6)
		valuesOffset++
		byte15 := int64(blocks[blocksOffset]) & 0xFF
		byte16 := int64(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte14 & 63) << 13) | (byte15 << 5) | (byte16 >> 3)
		byte17 := int64(blocks[blocksOffset]) & 0xFF
		byte18 := int64(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte16 & 7) << 16) | (byte17 << 8) | byte18
	}
}

func (b *BulkOperationPacked19) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := int32(blocks[blocksOffset]) & 0xFF
		byte1 := int32(blocks[blocksOffset]) & 0xFF
		byte2 := int32(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = (byte0 << 11) | (byte1 << 3) | (byte2 >> 5)
		valuesOffset++
		byte3 := int32(blocks[blocksOffset]) & 0xFF
		byte4 := int32(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte2 & 31) << 14) | (byte3 << 6) | (byte4 >> 2)
		valuesOffset++
		byte5 := int32(blocks[blocksOffset]) & 0xFF
		byte6 := int32(blocks[blocksOffset]) & 0xFF
		byte7 := int32(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte4 & 3) << 17) | (byte5 << 9) | (byte6 << 1) | (byte7 >> 7)
		valuesOffset++
		byte8 := int32(blocks[blocksOffset]) & 0xFF
		byte9 := int32(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte7 & 127) << 12) | (byte8 << 4) | (byte9 >> 4)
		valuesOffset++
		byte10 := int32(blocks[blocksOffset]) & 0xFF
		byte11 := int32(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte9 & 15) << 15) | (byte10 << 7) | (byte11 >> 1)
		valuesOffset++
		byte12 := int32(blocks[blocksOffset]) & 0xFF
		byte13 := int32(blocks[blocksOffset]) & 0xFF
		byte14 := int32(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte11 & 1) << 18) | (byte12 << 10) | (byte13 << 2) | (byte14 >> 6)
		valuesOffset++
		byte15 := int32(blocks[blocksOffset]) & 0xFF
		byte16 := int32(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte14 & 63) << 13) | (byte15 << 5) | (byte16 >> 3)
		byte17 := int32(blocks[blocksOffset]) & 0xFF
		byte18 := int32(blocks[blocksOffset]) & 0xFF
		values[valuesOffset] = ((byte16 & 7) << 16) | (byte17 << 8) | byte18
	}
}
