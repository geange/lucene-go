package packed

type BulkOperationPacked16 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked16() *BulkOperationPacked16 {
	return &BulkOperationPacked16{NewBulkOperationPacked(16)}
}

func (b *BulkOperationPacked16) DecodeLongToLong(blocks, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := blocks[blocksOffset]
		blocksOffset++
		for shift := 48; shift >= 0; shift -= 16 {
			values[valuesOffset] = (block >> shift) & 65535
			valuesOffset++
		}
	}
}

func (b *BulkOperationPacked16) DecodeByteToLong(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j++ {
		block0 := uint64(blocks[blocksOffset]) << 8
		block1 := uint64(blocks[blocksOffset+1] & 0xFF)
		values[valuesOffset] = block0 | block1
		valuesOffset++
		blocksOffset += 2
	}
}

func (b *BulkOperationPacked16) DecodeByteToInt(blocks []byte, values []uint32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j++ {
		block0 := uint32(blocks[blocksOffset]) << 8
		block1 := uint32(blocks[blocksOffset+1] & 0xFF)
		values[valuesOffset] = block0 | block1
		valuesOffset++
		blocksOffset += 2
	}
}
