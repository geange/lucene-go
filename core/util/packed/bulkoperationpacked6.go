package packed

type BulkOperationPacked6 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked6() *BulkOperationPacked6 {
	return &BulkOperationPacked6{NewBulkOperationPacked(6)}
}

func (b *BulkOperationPacked6) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
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

func (b *BulkOperationPacked6) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++

		values[valuesOffset] = byte0 >> 2
		valuesOffset++
		values[valuesOffset] = (byte0&3)<<4 | byte1>>4
		valuesOffset++
		values[valuesOffset] = (byte1&15)<<2 | byte2>>6
		valuesOffset++
		values[valuesOffset] = byte2 & 63
		valuesOffset++
	}
}

//func (b *BulkOperationPacked6) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
//	blocksOffset, valuesOffset := 0, 0
//	for i := 0; i < iterations; i++ {
//		byte0 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = byte0 >> 3
//		valuesOffset++
//		byte1 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = ((byte0 & 7) << 2) | (byte1 >> 6)
//		valuesOffset++
//		values[valuesOffset] = (byte1 >> 1) & 31
//		valuesOffset++
//		byte2 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = ((byte1 & 1) << 4) | (byte2 >> 4)
//		valuesOffset++
//		byte3 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = ((byte2 & 15) << 1) | (byte3 >> 7)
//		valuesOffset++
//		values[valuesOffset] = (byte3 >> 2) & 31
//		valuesOffset++
//		byte4 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = ((byte3 & 3) << 3) | (byte4 >> 5)
//		valuesOffset++
//		values[valuesOffset] = byte4 & 31
//		valuesOffset++
//	}
//}
