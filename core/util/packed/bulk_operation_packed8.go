package packed

type BulkOperationPacked8 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked8() *BulkOperationPacked8 {
	return &BulkOperationPacked8{NewBulkOperationPacked(8)}
}

func (b *BulkOperationPacked8) DecodeLongToLong(blocks, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := blocks[blocksOffset]
		blocksOffset++
		for shift := 56; shift >= 0; shift -= 8 {
			values[valuesOffset] = (block >> shift) & 255
			valuesOffset++
		}
	}
}

func (b *BulkOperationPacked8) DecodeByteToLong(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j++ {
		values[valuesOffset] = uint64(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		valuesOffset++
	}
}

func (b *BulkOperationPacked8) DecodeByteToInt(blocks []byte, values []uint32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j++ {
		values[valuesOffset] = uint32(blocks[blocksOffset] & 0xFF)
		blocksOffset++
		valuesOffset++
	}
}