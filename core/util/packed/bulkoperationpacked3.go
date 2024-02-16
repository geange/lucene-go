package packed

type BulkOperationPacked3 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked3() *BulkOperationPacked3 {
	return &BulkOperationPacked3{NewBulkOperationPacked(3)}
}

func (b *BulkOperationPacked3) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 61
		valuesOffset++
		values[valuesOffset] = (block0 >> 58) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 55) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 52) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 49) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 46) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 43) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 40) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 37) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 34) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 31) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 28) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 25) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 22) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 19) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 16) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 13) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 10) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 7) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 4) & 7
		valuesOffset++
		values[valuesOffset] = (block0 >> 1) & 7
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 1) << 2) | (block1 >> 62)
		valuesOffset++
		values[valuesOffset] = (block1 >> 59) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 56) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 53) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 50) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 47) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 44) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 41) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 38) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 35) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 32) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 29) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 26) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 23) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 20) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 17) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 14) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 11) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 8) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 5) & 7
		valuesOffset++
		values[valuesOffset] = (block1 >> 2) & 7
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 1) | (block2 >> 63)
		valuesOffset++
		values[valuesOffset] = (block2 >> 60) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 57) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 54) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 51) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 48) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 45) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 42) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 39) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 36) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 33) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 30) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 27) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 24) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 21) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 18) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 15) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 12) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 9) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 6) & 7
		valuesOffset++
		values[valuesOffset] = (block2 >> 3) & 7
		valuesOffset++
		values[valuesOffset] = block2 & 7
		valuesOffset++
	}
}

func (b *BulkOperationPacked3) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = byte0 >> 5
		valuesOffset++
		values[valuesOffset] = (byte0 >> 2) & 7
		valuesOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte0 & 3) << 1) | (byte1 >> 7)
		valuesOffset++
		values[valuesOffset] = (byte1 >> 4) & 7
		valuesOffset++
		values[valuesOffset] = (byte1 >> 1) & 7
		valuesOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte1 & 1) << 2) | (byte2 >> 6)
		valuesOffset++
		values[valuesOffset] = (byte2 >> 3) & 7
		valuesOffset++
		values[valuesOffset] = byte2 & 7
		valuesOffset++
	}
}

//func (b *BulkOperationPacked3) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
//	blocksOffset, valuesOffset := 0, 0
//	for i := 0; i < iterations; i++ {
//		byte0 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = byte0 >> 5
//		valuesOffset++
//		values[valuesOffset] = (byte0 >> 2) & 7
//		valuesOffset++
//		byte1 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = ((byte0 & 3) << 1) | (byte1 >> 7)
//		valuesOffset++
//		values[valuesOffset] = (byte1 >> 4) & 7
//		valuesOffset++
//		values[valuesOffset] = (byte1 >> 1) & 7
//		valuesOffset++
//		byte2 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = ((byte1 & 1) << 2) | (byte2 >> 6)
//		valuesOffset++
//		values[valuesOffset] = (byte2 >> 3) & 7
//		valuesOffset++
//		values[valuesOffset] = byte2 & 7
//		valuesOffset++
//	}
//}
