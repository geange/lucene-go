package packed

type BulkOperationPacked10 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked10() *BulkOperationPacked10 {
	return &BulkOperationPacked10{NewBulkOperationPacked(10)}
}

func (b *BulkOperationPacked10) DecodeLongToLong(blocks, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 54
		valuesOffset++
		values[valuesOffset] = (block0 >> 44) & 1023
		valuesOffset++
		values[valuesOffset] = (block0 >> 34) & 1023
		valuesOffset++
		values[valuesOffset] = (block0 >> 24) & 1023
		valuesOffset++
		values[valuesOffset] = (block0 >> 14) & 1023
		valuesOffset++
		values[valuesOffset] = (block0 >> 4) & 1023
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 15) << 6) | (block1 >> 58)
		valuesOffset++
		values[valuesOffset] = (block1 >> 48) & 1023
		valuesOffset++
		values[valuesOffset] = (block1 >> 38) & 1023
		valuesOffset++
		values[valuesOffset] = (block1 >> 28) & 1023
		valuesOffset++
		values[valuesOffset] = (block1 >> 18) & 1023
		valuesOffset++
		values[valuesOffset] = (block1 >> 8) & 1023
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 255) << 2) | (block2 >> 62)
		valuesOffset++
		values[valuesOffset] = (block2 >> 52) & 1023
		valuesOffset++
		values[valuesOffset] = (block2 >> 42) & 1023
		valuesOffset++
		values[valuesOffset] = (block2 >> 32) & 1023
		valuesOffset++
		values[valuesOffset] = (block2 >> 22) & 1023
		valuesOffset++
		values[valuesOffset] = (block2 >> 12) & 1023
		valuesOffset++
		values[valuesOffset] = (block2 >> 2) & 1023
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 3) << 8) | (block3 >> 56)
		valuesOffset++
		values[valuesOffset] = (block3 >> 46) & 1023
		valuesOffset++
		values[valuesOffset] = (block3 >> 36) & 1023
		valuesOffset++
		values[valuesOffset] = (block3 >> 26) & 1023
		valuesOffset++
		values[valuesOffset] = (block3 >> 16) & 1023
		valuesOffset++
		values[valuesOffset] = (block3 >> 6) & 1023
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 63) << 4) | (block4 >> 60)
		valuesOffset++
		values[valuesOffset] = (block4 >> 50) & 1023
		valuesOffset++
		values[valuesOffset] = (block4 >> 40) & 1023
		valuesOffset++
		values[valuesOffset] = (block4 >> 30) & 1023
		valuesOffset++
		values[valuesOffset] = (block4 >> 20) & 1023
		valuesOffset++
		values[valuesOffset] = (block4 >> 10) & 1023
		valuesOffset++
		values[valuesOffset] = block4 & 1023
		valuesOffset++
	}
}

func (b *BulkOperationPacked10) DecodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := int64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		byte1 := int64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = (byte0 << 2) | (byte1 >> 6)
		valuesOffset++
		byte2 := int64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = ((byte1 & 63) << 4) | (byte2 >> 4)
		valuesOffset++
		byte3 := int64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = ((byte2 & 15) << 6) | (byte3 >> 2)
		valuesOffset++
		byte4 := int64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = ((byte3 & 3) << 8) | byte4
		valuesOffset++
	}
}

func (b *BulkOperationPacked10) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := int32(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		byte1 := int32(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = (byte0 << 2) | (byte1 >> 6)
		valuesOffset++
		byte2 := int32(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = ((byte1 & 63) << 4) | (byte2 >> 4)
		valuesOffset++
		byte3 := int32(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = ((byte2 & 15) << 6) | (byte3 >> 2)
		valuesOffset++
		byte4 := int32(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = ((byte3 & 3) << 8) | byte4
		valuesOffset++
	}
}
