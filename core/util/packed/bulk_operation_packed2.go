package packed

type BulkOperationPacked2 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked2() *BulkOperationPacked2 {
	return &BulkOperationPacked2{NewBulkOperationPacked(2)}
}

func (b *BulkOperationPacked2) DecodeLongToLong(blocks, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := blocks[blocksOffset]
		blocksOffset++
		for shift := 62; shift >= 0; shift -= 2 {
			values[valuesOffset] = (block >> shift) & 3
			valuesOffset++
		}
	}
}

func (b *BulkOperationPacked2) DecodeByteToLong(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (block >> 6) & 3
		valuesOffset++
		values[valuesOffset] = (block >> 4) & 3
		valuesOffset++
		values[valuesOffset] = (block >> 2) & 3
		valuesOffset++
		values[valuesOffset] = block & 3
		valuesOffset++
	}
}

func (b *BulkOperationPacked2) DecodeByteToInt(blocks []byte, values []uint32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (block >> 6) & 3
		valuesOffset++
		values[valuesOffset] = (block >> 4) & 3
		valuesOffset++
		values[valuesOffset] = (block >> 2) & 3
		valuesOffset++
		values[valuesOffset] = block & 3
		valuesOffset++
	}
}