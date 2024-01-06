package packed

type BulkOperationPacked4 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked4() *BulkOperationPacked4 {
	return &BulkOperationPacked4{NewBulkOperationPacked(4)}
}

func (b *BulkOperationPacked4) DecodeLongToLong(blocks, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := blocks[blocksOffset]
		blocksOffset++
		for shift := 60; shift >= 0; shift -= 4 {
			values[valuesOffset] = (block >> shift) & 15
			valuesOffset++
		}
	}
}

func (b *BulkOperationPacked4) DecodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j++ {
		block := int64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (block >> 4) & 15
		valuesOffset++
		values[valuesOffset] = block & 15
		valuesOffset++
	}
}

func (b *BulkOperationPacked4) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j++ {
		block := int32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (block >> 4) & 15
		valuesOffset++
		values[valuesOffset] = block & 15
		valuesOffset++
	}
}
