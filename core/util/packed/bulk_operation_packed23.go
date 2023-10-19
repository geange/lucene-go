package packed

type BulkOperationPacked23 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked23() *BulkOperationPacked23 {
	return &BulkOperationPacked23{NewBulkOperationPacked(23)}
}

func (b *BulkOperationPacked23) DecodeLongToLong(blocks, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 41
		valuesOffset++
		values[valuesOffset] = (block0 >> 18) & 8388607
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 262143) << 5) | (block1 >> 59)
		valuesOffset++
		values[valuesOffset] = (block1 >> 36) & 8388607
		valuesOffset++
		values[valuesOffset] = (block1 >> 13) & 8388607
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 8191) << 10) | (block2 >> 54)
		valuesOffset++
		values[valuesOffset] = (block2 >> 31) & 8388607
		valuesOffset++
		values[valuesOffset] = (block2 >> 8) & 8388607
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 255) << 15) | (block3 >> 49)
		valuesOffset++
		values[valuesOffset] = (block3 >> 26) & 8388607
		valuesOffset++
		values[valuesOffset] = (block3 >> 3) & 8388607
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 7) << 20) | (block4 >> 44)
		valuesOffset++
		values[valuesOffset] = (block4 >> 21) & 8388607
		valuesOffset++
		block5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block4 & 2097151) << 2) | (block5 >> 62)
		valuesOffset++
		values[valuesOffset] = (block5 >> 39) & 8388607
		valuesOffset++
		values[valuesOffset] = (block5 >> 16) & 8388607
		valuesOffset++
		block6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block5 & 65535) << 7) | (block6 >> 57)
		valuesOffset++
		values[valuesOffset] = (block6 >> 34) & 8388607
		valuesOffset++
		values[valuesOffset] = (block6 >> 11) & 8388607
		valuesOffset++
		block7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block6 & 2047) << 12) | (block7 >> 52)
		valuesOffset++
		values[valuesOffset] = (block7 >> 29) & 8388607
		valuesOffset++
		values[valuesOffset] = (block7 >> 6) & 8388607
		valuesOffset++
		block8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block7 & 63) << 17) | (block8 >> 47)
		valuesOffset++
		values[valuesOffset] = (block8 >> 24) & 8388607
		valuesOffset++
		values[valuesOffset] = (block8 >> 1) & 8388607
		valuesOffset++
		block9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block8 & 1) << 22) | (block9 >> 42)
		valuesOffset++
		values[valuesOffset] = (block9 >> 19) & 8388607
		valuesOffset++
		block10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block9 & 524287) << 4) | (block10 >> 60)
		valuesOffset++
		values[valuesOffset] = (block10 >> 37) & 8388607
		valuesOffset++
		values[valuesOffset] = (block10 >> 14) & 8388607
		valuesOffset++
		block11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block10 & 16383) << 9) | (block11 >> 55)
		valuesOffset++
		values[valuesOffset] = (block11 >> 32) & 8388607
		valuesOffset++
		values[valuesOffset] = (block11 >> 9) & 8388607
		valuesOffset++
		block12 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block11 & 511) << 14) | (block12 >> 50)
		valuesOffset++
		values[valuesOffset] = (block12 >> 27) & 8388607
		valuesOffset++
		values[valuesOffset] = (block12 >> 4) & 8388607
		valuesOffset++
		block13 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block12 & 15) << 19) | (block13 >> 45)
		valuesOffset++
		values[valuesOffset] = (block13 >> 22) & 8388607
		valuesOffset++
		block14 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block13 & 4194303) << 1) | (block14 >> 63)
		valuesOffset++
		values[valuesOffset] = (block14 >> 40) & 8388607
		valuesOffset++
		values[valuesOffset] = (block14 >> 17) & 8388607
		valuesOffset++
		block15 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block14 & 131071) << 6) | (block15 >> 58)
		valuesOffset++
		values[valuesOffset] = (block15 >> 35) & 8388607
		valuesOffset++
		values[valuesOffset] = (block15 >> 12) & 8388607
		valuesOffset++
		block16 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block15 & 4095) << 11) | (block16 >> 53)
		valuesOffset++
		values[valuesOffset] = (block16 >> 30) & 8388607
		valuesOffset++
		values[valuesOffset] = (block16 >> 7) & 8388607
		valuesOffset++
		block17 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block16 & 127) << 16) | (block17 >> 48)
		valuesOffset++
		values[valuesOffset] = (block17 >> 25) & 8388607
		valuesOffset++
		values[valuesOffset] = (block17 >> 2) & 8388607
		valuesOffset++
		block18 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block17 & 3) << 21) | (block18 >> 43)
		valuesOffset++
		values[valuesOffset] = (block18 >> 20) & 8388607
		valuesOffset++
		block19 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block18 & 1048575) << 3) | (block19 >> 61)
		valuesOffset++
		values[valuesOffset] = (block19 >> 38) & 8388607
		valuesOffset++
		values[valuesOffset] = (block19 >> 15) & 8388607
		valuesOffset++
		block20 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block19 & 32767) << 8) | (block20 >> 56)
		valuesOffset++
		values[valuesOffset] = (block20 >> 33) & 8388607
		valuesOffset++
		values[valuesOffset] = (block20 >> 10) & 8388607
		valuesOffset++
		block21 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block20 & 1023) << 13) | (block21 >> 51)
		valuesOffset++
		values[valuesOffset] = (block21 >> 28) & 8388607
		valuesOffset++
		values[valuesOffset] = (block21 >> 5) & 8388607
		valuesOffset++
		block22 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block21 & 31) << 18) | (block22 >> 46)
		valuesOffset++
		values[valuesOffset] = (block22 >> 23) & 8388607
		valuesOffset++
		values[valuesOffset] = block22 & 8388607
		valuesOffset++
	}
}

func (b *BulkOperationPacked23) DecodeByteToLong(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 15) | (byte1 << 7) | (byte2 >> 1)
		valuesOffset++

		byte3 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte4 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte5 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte2 & 1) << 22) | (byte3 << 14) | (byte4 << 6) | (byte5 >> 2)
		valuesOffset++
		byte6 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte7 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte8 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte5 & 3) << 21) | (byte6 << 13) | (byte7 << 5) | (byte8 >> 3)
		valuesOffset++
		byte9 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte10 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte11 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte8 & 7) << 20) | (byte9 << 12) | (byte10 << 4) | (byte11 >> 4)
		valuesOffset++
		byte12 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte13 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte14 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte11 & 15) << 19) | (byte12 << 11) | (byte13 << 3) | (byte14 >> 5)
		valuesOffset++
		byte15 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte16 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte17 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte14 & 31) << 18) | (byte15 << 10) | (byte16 << 2) | (byte17 >> 6)
		valuesOffset++
		byte18 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte19 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte20 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte17 & 63) << 17) | (byte18 << 9) | (byte19 << 1) | (byte20 >> 7)
		valuesOffset++
		byte21 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte22 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte20 & 127) << 16) | (byte21 << 8) | byte22
		valuesOffset++
	}
}

func (b *BulkOperationPacked23) DecodeByteToInt(blocks []byte, values []uint32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte2 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 15) | (byte1 << 7) | (byte2 >> 1)
		valuesOffset++

		byte3 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte4 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte5 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte2 & 1) << 22) | (byte3 << 14) | (byte4 << 6) | (byte5 >> 2)
		valuesOffset++
		byte6 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte7 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte8 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte5 & 3) << 21) | (byte6 << 13) | (byte7 << 5) | (byte8 >> 3)
		valuesOffset++
		byte9 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte10 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte11 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte8 & 7) << 20) | (byte9 << 12) | (byte10 << 4) | (byte11 >> 4)
		valuesOffset++
		byte12 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte13 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte14 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte11 & 15) << 19) | (byte12 << 11) | (byte13 << 3) | (byte14 >> 5)
		valuesOffset++
		byte15 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte16 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte17 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte14 & 31) << 18) | (byte15 << 10) | (byte16 << 2) | (byte17 >> 6)
		valuesOffset++
		byte18 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte19 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte20 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte17 & 63) << 17) | (byte18 << 9) | (byte19 << 1) | (byte20 >> 7)
		valuesOffset++
		byte21 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte22 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte20 & 127) << 16) | (byte21 << 8) | byte22
		valuesOffset++
	}
}
