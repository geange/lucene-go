package packed

type BulkOperationPacked24 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked24() *BulkOperationPacked24 {
	return &BulkOperationPacked24{NewBulkOperationPacked(24)}
}

func (b *BulkOperationPacked24) DecodeLongToLong(blocks, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 40
		valuesOffset++
		values[valuesOffset] = (block0 >> 16) & 16777215
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 65535) << 8) | (block1 >> 56)
		valuesOffset++
		values[valuesOffset] = (block1 >> 32) & 16777215
		valuesOffset++
		values[valuesOffset] = (block1 >> 8) & 16777215
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 255) << 16) | (block2 >> 48)
		valuesOffset++
		values[valuesOffset] = (block2 >> 24) & 16777215
		valuesOffset++
		values[valuesOffset] = block2 & 16777215
		valuesOffset++
	}
}

func (b *BulkOperationPacked24) DecodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte1 := int64((blocks[blocksOffset]) & 0xFF)
		blocksOffset++
		byte2 := int64(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = (byte0 << 16) | (byte1 << 8) | byte2
		valuesOffset++
	}
}

func (b *BulkOperationPacked24) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		byte1 := int((blocks[blocksOffset]) & 0xFF)
		blocksOffset++
		byte2 := int32(blocks[blocksOffset]) & 0xFF
		blocksOffset++
		values[valuesOffset] = (byte0 << 16) | (byte1 << 8) | byte2
		valuesOffset++
	}
}
