package packed

type BulkOperationPacked24 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked24() *BulkOperationPacked24 {
	return &BulkOperationPacked24{NewBulkOperationPacked(24)}
}

func (b *BulkOperationPacked24) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 40
		valuesOffset++
		values[valuesOffset] = (block0 >> 16) & 0xFFFFFF
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 65535) << 8) | (block1 >> 56)
		valuesOffset++
		values[valuesOffset] = (block1 >> 32) & 0xFFFFFF
		valuesOffset++
		values[valuesOffset] = (block1 >> 8) & 0xFFFFFF
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 255) << 16) | (block2 >> 48)
		valuesOffset++
		values[valuesOffset] = (block2 >> 24) & 0xFFFFFF
		valuesOffset++
		values[valuesOffset] = block2 & 0xFFFFFF
		valuesOffset++
	}
}

func (b *BulkOperationPacked24) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 16) | (byte1 << 8) | byte2
		valuesOffset++
	}
}
