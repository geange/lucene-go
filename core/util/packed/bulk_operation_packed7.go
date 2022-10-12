package packed

type BulkOperationPacked7 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked7() *BulkOperationPacked7 {
	return &BulkOperationPacked7{NewBulkOperationPacked(7)}
}

func (b *BulkOperationPacked7) DecodeLongToLong(blocks, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 58
		valuesOffset++
		values[valuesOffset] = (block0 >> 52) & 63
		valuesOffset++
		values[valuesOffset] = (block0 >> 46) & 63
		valuesOffset++
		values[valuesOffset] = (block0 >> 40) & 63
		valuesOffset++
		values[valuesOffset] = (block0 >> 34) & 63
		valuesOffset++
		values[valuesOffset] = (block0 >> 28) & 63
		valuesOffset++
		values[valuesOffset] = (block0 >> 22) & 63
		valuesOffset++
		values[valuesOffset] = (block0 >> 16) & 63
		valuesOffset++
		values[valuesOffset] = (block0 >> 10) & 63
		valuesOffset++
		values[valuesOffset] = (block0 >> 4) & 63
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 15) << 2) | (block1 >> 62)
		valuesOffset++
		values[valuesOffset] = (block1 >> 56) & 63
		valuesOffset++
		values[valuesOffset] = (block1 >> 50) & 63
		valuesOffset++
		values[valuesOffset] = (block1 >> 44) & 63
		valuesOffset++
		values[valuesOffset] = (block1 >> 38) & 63
		valuesOffset++
		values[valuesOffset] = (block1 >> 32) & 63
		valuesOffset++
		values[valuesOffset] = (block1 >> 26) & 63
		valuesOffset++
		values[valuesOffset] = (block1 >> 20) & 63
		valuesOffset++
		values[valuesOffset] = (block1 >> 14) & 63
		valuesOffset++
		values[valuesOffset] = (block1 >> 8) & 63
		valuesOffset++
		values[valuesOffset] = (block1 >> 2) & 63
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 4) | (block2 >> 60)
		valuesOffset++
		values[valuesOffset] = (block2 >> 54) & 63
		valuesOffset++
		values[valuesOffset] = (block2 >> 48) & 63
		valuesOffset++
		values[valuesOffset] = (block2 >> 42) & 63
		valuesOffset++
		values[valuesOffset] = (block2 >> 36) & 63
		valuesOffset++
		values[valuesOffset] = (block2 >> 30) & 63
		valuesOffset++
		values[valuesOffset] = (block2 >> 24) & 63
		valuesOffset++
		values[valuesOffset] = (block2 >> 18) & 63
		valuesOffset++
		values[valuesOffset] = (block2 >> 12) & 63
		valuesOffset++
		values[valuesOffset] = (block2 >> 6) & 63
		valuesOffset++
		values[valuesOffset] = block2 & 63
		valuesOffset++
	}
}

func (b *BulkOperationPacked7) DecodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := int64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = byte0 >> 2
		valuesOffset++
		byte1 := int64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = ((byte0 & 3) << 4) | (byte1 >> 4)
		valuesOffset++
		byte2 := int64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = ((byte1 & 15) << 2) | (byte2 >> 6)
		valuesOffset++
		values[valuesOffset] = byte2 & 63
		valuesOffset++
	}
}

func (b *BulkOperationPacked7) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := int32(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = byte0 >> 2
		valuesOffset++
		byte1 := int32(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = ((byte0 & 3) << 4) | (byte1 >> 4)
		valuesOffset++
		byte2 := int32(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		values[valuesOffset] = ((byte1 & 15) << 2) | (byte2 >> 6)
		valuesOffset++
		values[valuesOffset] = byte2 & 63
		valuesOffset++
	}
}
