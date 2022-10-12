package packed

type BulkOperationPacked20 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked20() *BulkOperationPacked20 {
	return &BulkOperationPacked20{NewBulkOperationPacked(20)}
}

func (b *BulkOperationPacked20) DecodeLongToLong(blocks, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 44
		valuesOffset++
		values[valuesOffset] = (block0 >> 24) & 1048575
		valuesOffset++
		values[valuesOffset] = (block0 >> 4) & 1048575
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 15) << 16) | (block1 >> 48)
		valuesOffset++
		values[valuesOffset] = (block1 >> 28) & 1048575
		valuesOffset++
		values[valuesOffset] = (block1 >> 8) & 1048575
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 255) << 12) | (block2 >> 52)
		valuesOffset++
		values[valuesOffset] = (block2 >> 32) & 1048575
		valuesOffset++
		values[valuesOffset] = (block2 >> 12) & 1048575
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 4095) << 8) | (block3 >> 56)
		valuesOffset++
		values[valuesOffset] = (block3 >> 36) & 1048575
		valuesOffset++
		values[valuesOffset] = (block3 >> 16) & 1048575
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 65535) << 4) | (block4 >> 60)
		valuesOffset++
		values[valuesOffset] = (block4 >> 40) & 1048575
		valuesOffset++
		values[valuesOffset] = (block4 >> 20) & 1048575
		valuesOffset++
		values[valuesOffset] = block4 & 1048575
		valuesOffset++
	}
}

func (b *BulkOperationPacked20) DecodeByteToLong(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		byte2 := uint64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = (byte0 << 12) | (byte1 << 4) | (byte2 >> 4)
		valuesOffset++
		byte3 := uint64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		byte4 := uint64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = ((byte2 & 15) << 16) | (byte3 << 8) | byte4
		valuesOffset++
	}
}

func (b *BulkOperationPacked20) DecodeByteToInt(blocks []byte, values []uint32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte1 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte2 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = (byte0 << 12) | (byte1 << 4) | (byte2 >> 4)
		valuesOffset++
		byte3 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte4 := uint32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = ((byte2 & 15) << 16) | (byte3 << 8) | byte4
		valuesOffset++
	}
}
