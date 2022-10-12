package packed

type BulkOperationPacked18 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked18() *BulkOperationPacked18 {
	return &BulkOperationPacked18{NewBulkOperationPacked(18)}
}

func (b *BulkOperationPacked18) DecodeLongToLong(blocks, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 46
		valuesOffset++
		values[valuesOffset] = (block0 >> 28) & 262143
		valuesOffset++
		values[valuesOffset] = (block0 >> 10) & 262143
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 1023) << 8) | (block1 >> 56)
		valuesOffset++
		values[valuesOffset] = (block1 >> 38) & 262143
		valuesOffset++
		values[valuesOffset] = (block1 >> 20) & 262143
		valuesOffset++
		values[valuesOffset] = (block1 >> 2) & 262143
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 16) | (block2 >> 48)
		valuesOffset++
		values[valuesOffset] = (block2 >> 30) & 262143
		valuesOffset++
		values[valuesOffset] = (block2 >> 12) & 262143
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 4095) << 6) | (block3 >> 58)
		valuesOffset++
		values[valuesOffset] = (block3 >> 40) & 262143
		valuesOffset++
		values[valuesOffset] = (block3 >> 22) & 262143
		valuesOffset++
		values[valuesOffset] = (block3 >> 4) & 262143
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 15) << 14) | (block4 >> 50)
		valuesOffset++
		values[valuesOffset] = (block4 >> 32) & 262143
		valuesOffset++
		values[valuesOffset] = (block4 >> 14) & 262143
		valuesOffset++
		block5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block4 & 16383) << 4) | (block5 >> 60)
		valuesOffset++
		values[valuesOffset] = (block5 >> 42) & 262143
		valuesOffset++
		values[valuesOffset] = (block5 >> 24) & 262143
		valuesOffset++
		values[valuesOffset] = (block5 >> 6) & 262143
		valuesOffset++
		block6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block5 & 63) << 12) | (block6 >> 52)
		valuesOffset++
		values[valuesOffset] = (block6 >> 34) & 262143
		valuesOffset++
		values[valuesOffset] = (block6 >> 16) & 262143
		valuesOffset++
		block7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block6 & 65535) << 2) | (block7 >> 62)
		valuesOffset++
		values[valuesOffset] = (block7 >> 44) & 262143
		valuesOffset++
		values[valuesOffset] = (block7 >> 26) & 262143
		valuesOffset++
		values[valuesOffset] = (block7 >> 8) & 262143
		valuesOffset++
		block8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block7 & 255) << 10) | (block8 >> 54)
		valuesOffset++
		values[valuesOffset] = (block8 >> 36) & 262143
		valuesOffset++
		values[valuesOffset] = (block8 >> 18) & 262143
		valuesOffset++
		values[valuesOffset] = block8 & 262143
		valuesOffset++
	}
}

func (b *BulkOperationPacked18) DecodeByteToLong(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte2 := uint64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = (byte0 << 10) | (byte1 << 2) | (byte2 >> 6)
		valuesOffset++
		byte3 := uint64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte4 := uint64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte2 & 63) << 12) | (byte3 << 4) | (byte4 >> 4)
		valuesOffset++
		byte5 := uint64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte6 := uint64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte4 & 15) << 14) | (byte5 << 6) | (byte6 >> 2)
		valuesOffset++
		byte7 := uint64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte8 := uint64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte6 & 3) << 16) | (byte7 << 8) | byte8
		valuesOffset++
	}
}

func (b *BulkOperationPacked18) DecodeByteToInt(blocks []byte, values []uint32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte1 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte2 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = (byte0 << 10) | (byte1 << 2) | (byte2 >> 6)
		valuesOffset++
		byte3 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte4 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte2 & 63) << 12) | (byte3 << 4) | (byte4 >> 4)
		valuesOffset++
		byte5 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte6 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte4 & 15) << 14) | (byte5 << 6) | (byte6 >> 2)
		valuesOffset++
		byte7 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte8 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte6 & 3) << 16) | (byte7 << 8) | byte8
		valuesOffset++
	}
}
