package packed

type BulkOperationPacked11 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked11() *BulkOperationPacked11 {
	return &BulkOperationPacked11{NewBulkOperationPacked(11)}
}

func (b *BulkOperationPacked11) DecodeLongToLong(blocks, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 53
		valuesOffset++
		values[valuesOffset] = (block0 >> 42) & 2047
		valuesOffset++
		values[valuesOffset] = (block0 >> 31) & 2047
		valuesOffset++
		values[valuesOffset] = (block0 >> 20) & 2047
		valuesOffset++
		values[valuesOffset] = (block0 >> 9) & 2047
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 511) << 2) | (block1 >> 62)
		valuesOffset++
		values[valuesOffset] = (block1 >> 51) & 2047
		valuesOffset++
		values[valuesOffset] = (block1 >> 40) & 2047
		valuesOffset++
		values[valuesOffset] = (block1 >> 29) & 2047
		valuesOffset++
		values[valuesOffset] = (block1 >> 18) & 2047
		valuesOffset++
		values[valuesOffset] = (block1 >> 7) & 2047
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 127) << 4) | (block2 >> 60)
		valuesOffset++
		values[valuesOffset] = (block2 >> 49) & 2047
		valuesOffset++
		values[valuesOffset] = (block2 >> 38) & 2047
		valuesOffset++
		values[valuesOffset] = (block2 >> 27) & 2047
		valuesOffset++
		values[valuesOffset] = (block2 >> 16) & 2047
		valuesOffset++
		values[valuesOffset] = (block2 >> 5) & 2047
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 31) << 6) | (block3 >> 58)
		valuesOffset++
		values[valuesOffset] = (block3 >> 47) & 2047
		valuesOffset++
		values[valuesOffset] = (block3 >> 36) & 2047
		valuesOffset++
		values[valuesOffset] = (block3 >> 25) & 2047
		valuesOffset++
		values[valuesOffset] = (block3 >> 14) & 2047
		valuesOffset++
		values[valuesOffset] = (block3 >> 3) & 2047
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 7) << 8) | (block4 >> 56)
		valuesOffset++
		values[valuesOffset] = (block4 >> 45) & 2047
		valuesOffset++
		values[valuesOffset] = (block4 >> 34) & 2047
		valuesOffset++
		values[valuesOffset] = (block4 >> 23) & 2047
		valuesOffset++
		values[valuesOffset] = (block4 >> 12) & 2047
		valuesOffset++
		values[valuesOffset] = (block4 >> 1) & 2047
		valuesOffset++
		block5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block4 & 1) << 10) | (block5 >> 54)
		valuesOffset++
		values[valuesOffset] = (block5 >> 43) & 2047
		valuesOffset++
		values[valuesOffset] = (block5 >> 32) & 2047
		valuesOffset++
		values[valuesOffset] = (block5 >> 21) & 2047
		valuesOffset++
		values[valuesOffset] = (block5 >> 10) & 2047
		valuesOffset++
		block6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block5 & 1023) << 1) | (block6 >> 63)
		valuesOffset++
		values[valuesOffset] = (block6 >> 52) & 2047
		valuesOffset++
		values[valuesOffset] = (block6 >> 41) & 2047
		valuesOffset++
		values[valuesOffset] = (block6 >> 30) & 2047
		valuesOffset++
		values[valuesOffset] = (block6 >> 19) & 2047
		valuesOffset++
		values[valuesOffset] = (block6 >> 8) & 2047
		valuesOffset++
		block7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block6 & 255) << 3) | (block7 >> 61)
		valuesOffset++
		values[valuesOffset] = (block7 >> 50) & 2047
		valuesOffset++
		values[valuesOffset] = (block7 >> 39) & 2047
		valuesOffset++
		values[valuesOffset] = (block7 >> 28) & 2047
		valuesOffset++
		values[valuesOffset] = (block7 >> 17) & 2047
		valuesOffset++
		values[valuesOffset] = (block7 >> 6) & 2047
		valuesOffset++
		block8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block7 & 63) << 5) | (block8 >> 59)
		valuesOffset++
		values[valuesOffset] = (block8 >> 48) & 2047
		valuesOffset++
		values[valuesOffset] = (block8 >> 37) & 2047
		valuesOffset++
		values[valuesOffset] = (block8 >> 26) & 2047
		valuesOffset++
		values[valuesOffset] = (block8 >> 15) & 2047
		valuesOffset++
		values[valuesOffset] = (block8 >> 4) & 2047
		valuesOffset++
		block9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block8 & 15) << 7) | (block9 >> 57)
		valuesOffset++
		values[valuesOffset] = (block9 >> 46) & 2047
		valuesOffset++
		values[valuesOffset] = (block9 >> 35) & 2047
		valuesOffset++
		values[valuesOffset] = (block9 >> 24) & 2047
		valuesOffset++
		values[valuesOffset] = (block9 >> 13) & 2047
		valuesOffset++
		values[valuesOffset] = (block9 >> 2) & 2047
		valuesOffset++
		block10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block9 & 3) << 9) | (block10 >> 55)
		valuesOffset++
		values[valuesOffset] = (block10 >> 44) & 2047
		valuesOffset++
		values[valuesOffset] = (block10 >> 33) & 2047
		valuesOffset++
		values[valuesOffset] = (block10 >> 22) & 2047
		valuesOffset++
		values[valuesOffset] = (block10 >> 11) & 2047
		valuesOffset++
		values[valuesOffset] = block10 & 2047
		valuesOffset++
	}
}

func (b *BulkOperationPacked11) DecodeByteToLong(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 3) | (byte1 >> 5)
		valuesOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte1 & 31) << 6) | (byte2 >> 2)
		valuesOffset++
		byte3 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte4 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte2 & 3) << 9) | (byte3 << 1) | (byte4 >> 7)
		valuesOffset++
		byte5 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte4 & 127) << 4) | (byte5 >> 4)
		valuesOffset++
		byte6 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte5 & 15) << 7) | (byte6 >> 1)
		valuesOffset++
		byte7 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte8 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte6 & 1) << 10) | (byte7 << 2) | (byte8 >> 6)
		valuesOffset++
		byte9 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte8 & 63) << 5) | (byte9 >> 3)
		valuesOffset++
		byte10 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte9 & 7) << 8) | byte10
		valuesOffset++
	}
}

func (b *BulkOperationPacked11) DecodeByteToInt(blocks []byte, values []uint32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 3) | (byte1 >> 5)
		valuesOffset++
		byte2 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte1 & 31) << 6) | (byte2 >> 2)
		valuesOffset++
		byte3 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte4 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte2 & 3) << 9) | (byte3 << 1) | (byte4 >> 7)
		valuesOffset++
		byte5 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte4 & 127) << 4) | (byte5 >> 4)
		valuesOffset++
		byte6 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte5 & 15) << 7) | (byte6 >> 1)
		valuesOffset++
		byte7 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte8 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte6 & 1) << 10) | (byte7 << 2) | (byte8 >> 6)
		valuesOffset++
		byte9 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte8 & 63) << 5) | (byte9 >> 3)
		valuesOffset++
		byte10 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte9 & 7) << 8) | byte10
		valuesOffset++
	}
}
