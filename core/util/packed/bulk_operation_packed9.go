package packed

type BulkOperationPacked9 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked9() *BulkOperationPacked9 {
	return &BulkOperationPacked9{NewBulkOperationPacked(9)}
}

func (b *BulkOperationPacked9) DecodeLongToLong(blocks, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 55
		valuesOffset++
		values[valuesOffset] = (block0 >> 46) & 511
		valuesOffset++
		values[valuesOffset] = (block0 >> 37) & 511
		valuesOffset++
		values[valuesOffset] = (block0 >> 28) & 511
		valuesOffset++
		values[valuesOffset] = (block0 >> 19) & 511
		valuesOffset++
		values[valuesOffset] = (block0 >> 10) & 511
		valuesOffset++
		values[valuesOffset] = (block0 >> 1) & 511
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 1) << 8) | (block1 >> 56)
		valuesOffset++
		values[valuesOffset] = (block1 >> 47) & 511
		valuesOffset++
		values[valuesOffset] = (block1 >> 38) & 511
		valuesOffset++
		values[valuesOffset] = (block1 >> 29) & 511
		valuesOffset++
		values[valuesOffset] = (block1 >> 20) & 511
		valuesOffset++
		values[valuesOffset] = (block1 >> 11) & 511
		valuesOffset++
		values[valuesOffset] = (block1 >> 2) & 511
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 7) | (block2 >> 57)
		valuesOffset++
		values[valuesOffset] = (block2 >> 48) & 511
		valuesOffset++
		values[valuesOffset] = (block2 >> 39) & 511
		valuesOffset++
		values[valuesOffset] = (block2 >> 30) & 511
		valuesOffset++
		values[valuesOffset] = (block2 >> 21) & 511
		valuesOffset++
		values[valuesOffset] = (block2 >> 12) & 511
		valuesOffset++
		values[valuesOffset] = (block2 >> 3) & 511
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 7) << 6) | (block3 >> 58)
		valuesOffset++
		values[valuesOffset] = (block3 >> 49) & 511
		valuesOffset++
		values[valuesOffset] = (block3 >> 40) & 511
		valuesOffset++
		values[valuesOffset] = (block3 >> 31) & 511
		valuesOffset++
		values[valuesOffset] = (block3 >> 22) & 511
		valuesOffset++
		values[valuesOffset] = (block3 >> 13) & 511
		valuesOffset++
		values[valuesOffset] = (block3 >> 4) & 511
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 15) << 5) | (block4 >> 59)
		valuesOffset++
		values[valuesOffset] = (block4 >> 50) & 511
		valuesOffset++
		values[valuesOffset] = (block4 >> 41) & 511
		valuesOffset++
		values[valuesOffset] = (block4 >> 32) & 511
		valuesOffset++
		values[valuesOffset] = (block4 >> 23) & 511
		valuesOffset++
		values[valuesOffset] = (block4 >> 14) & 511
		valuesOffset++
		values[valuesOffset] = (block4 >> 5) & 511
		valuesOffset++
		block5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block4 & 31) << 4) | (block5 >> 60)
		valuesOffset++
		values[valuesOffset] = (block5 >> 51) & 511
		valuesOffset++
		values[valuesOffset] = (block5 >> 42) & 511
		valuesOffset++
		values[valuesOffset] = (block5 >> 33) & 511
		valuesOffset++
		values[valuesOffset] = (block5 >> 24) & 511
		valuesOffset++
		values[valuesOffset] = (block5 >> 15) & 511
		valuesOffset++
		values[valuesOffset] = (block5 >> 6) & 511
		valuesOffset++
		block6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block5 & 63) << 3) | (block6 >> 61)
		valuesOffset++
		values[valuesOffset] = (block6 >> 52) & 511
		valuesOffset++
		values[valuesOffset] = (block6 >> 43) & 511
		valuesOffset++
		values[valuesOffset] = (block6 >> 34) & 511
		valuesOffset++
		values[valuesOffset] = (block6 >> 25) & 511
		valuesOffset++
		values[valuesOffset] = (block6 >> 16) & 511
		valuesOffset++
		values[valuesOffset] = (block6 >> 7) & 511
		valuesOffset++
		block7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block6 & 127) << 2) | (block7 >> 62)
		valuesOffset++
		values[valuesOffset] = (block7 >> 53) & 511
		valuesOffset++
		values[valuesOffset] = (block7 >> 44) & 511
		valuesOffset++
		values[valuesOffset] = (block7 >> 35) & 511
		valuesOffset++
		values[valuesOffset] = (block7 >> 26) & 511
		valuesOffset++
		values[valuesOffset] = (block7 >> 17) & 511
		valuesOffset++
		values[valuesOffset] = (block7 >> 8) & 511
		valuesOffset++
		block8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block7 & 255) << 1) | (block8 >> 63)
		valuesOffset++
		values[valuesOffset] = (block8 >> 54) & 511
		valuesOffset++
		values[valuesOffset] = (block8 >> 45) & 511
		valuesOffset++
		values[valuesOffset] = (block8 >> 36) & 511
		valuesOffset++
		values[valuesOffset] = (block8 >> 27) & 511
		valuesOffset++
		values[valuesOffset] = (block8 >> 18) & 511
		valuesOffset++
		values[valuesOffset] = (block8 >> 9) & 511
		valuesOffset++
		values[valuesOffset] = block8 & 511
		valuesOffset++
	}
}

func (b *BulkOperationPacked9) DecodeByteToLong(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 1) | (byte1 >> 7)
		valuesOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte1 & 127) << 2) | (byte2 >> 6)
		valuesOffset++
		byte3 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte2 & 63) << 3) | (byte3 >> 5)
		valuesOffset++
		byte4 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte3 & 31) << 4) | (byte4 >> 4)
		valuesOffset++
		byte5 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte4 & 15) << 5) | (byte5 >> 3)
		valuesOffset++
		byte6 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte5 & 7) << 6) | (byte6 >> 2)
		valuesOffset++
		byte7 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte6 & 3) << 7) | (byte7 >> 1)
		valuesOffset++
		byte8 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte7 & 1) << 8) | byte8
		valuesOffset++
	}
}

func (b *BulkOperationPacked9) DecodeByteToInt(blocks []byte, values []uint32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 1) | (byte1 >> 7)
		valuesOffset++
		byte2 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte1 & 127) << 2) | (byte2 >> 6)
		valuesOffset++
		byte3 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte2 & 63) << 3) | (byte3 >> 5)
		valuesOffset++
		byte4 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte3 & 31) << 4) | (byte4 >> 4)
		valuesOffset++
		byte5 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte4 & 15) << 5) | (byte5 >> 3)
		valuesOffset++
		byte6 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte5 & 7) << 6) | (byte6 >> 2)
		valuesOffset++
		byte7 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte6 & 3) << 7) | (byte7 >> 1)
		valuesOffset++
		byte8 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte7 & 1) << 8) | byte8
		valuesOffset++
	}
}
