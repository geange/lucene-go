package packed

type BulkOperationPacked1 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked1() *BulkOperationPacked1 {
	return &BulkOperationPacked1{NewBulkOperationPacked(1)}
}

func (b *BulkOperationPacked1) DecodeLongToLong(blocks, values []int64, iterations int) {
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

func (b *BulkOperationPacked1) DecodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := int64(blocks[blocksOffset])
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

//func (b *BulkOperationPacked1) DecodeLongToInt(blocks []int64, values []int, iterations int) {
//	blocksOffset, valuesOffset := 0, 0
//	for i := 0; i < iterations; i++ {
//		block := blocks[blocksOffset]
//		blocksOffset++
//		for shift := 63; shift >= 0; shift -= 1 {
//			values[valuesOffset] = (block >> shift) & 1
//			valuesOffset++
//		}
//	}
//}

func (b *BulkOperationPacked1) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := int32(blocks[blocksOffset])
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
