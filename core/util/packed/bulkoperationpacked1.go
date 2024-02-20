package packed

type BulkOperationPacked1 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked1() *BulkOperationPacked1 {
	return &BulkOperationPacked1{NewBulkOperationPacked(1)}
}

func (b *BulkOperationPacked1) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := blocks[blocksOffset]
		blocksOffset++
		for shift := 63; shift >= 0; shift -= 1 {
			values[valuesOffset] = (block >> shift) & 1
			valuesOffset++
		}
	}
}

func (b *BulkOperationPacked1) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (block >> 7) & 1
		valuesOffset++
		values[valuesOffset] = (block >> 6) & 1
		valuesOffset++
		values[valuesOffset] = (block >> 5) & 1
		valuesOffset++
		values[valuesOffset] = (block >> 4) & 1
		valuesOffset++
		values[valuesOffset] = (block >> 3) & 1
		valuesOffset++
		values[valuesOffset] = (block >> 2) & 1
		valuesOffset++
		values[valuesOffset] = (block >> 1) & 1
		valuesOffset++
		values[valuesOffset] = block & 1
		valuesOffset++
	}
}
