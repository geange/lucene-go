package packed

type BulkOperationPacked14 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked14() *BulkOperationPacked14 {
	return &BulkOperationPacked14{NewBulkOperationPacked(14)}
}

func (b *BulkOperationPacked14) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 50
		valuesOffset++
		values[valuesOffset] = (block0 >> 36) & 16383
		valuesOffset++
		values[valuesOffset] = (block0 >> 22) & 16383
		valuesOffset++
		values[valuesOffset] = (block0 >> 8) & 16383
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 255) << 6) | (block1 >> 58)
		valuesOffset++
		values[valuesOffset] = (block1 >> 44) & 16383
		valuesOffset++
		values[valuesOffset] = (block1 >> 30) & 16383
		valuesOffset++
		values[valuesOffset] = (block1 >> 16) & 16383
		valuesOffset++
		values[valuesOffset] = (block1 >> 2) & 16383
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 12) | (block2 >> 52)
		valuesOffset++
		values[valuesOffset] = (block2 >> 38) & 16383
		valuesOffset++
		values[valuesOffset] = (block2 >> 24) & 16383
		valuesOffset++
		values[valuesOffset] = (block2 >> 10) & 16383
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 1023) << 4) | (block3 >> 60)
		valuesOffset++
		values[valuesOffset] = (block3 >> 46) & 16383
		valuesOffset++
		values[valuesOffset] = (block3 >> 32) & 16383
		valuesOffset++
		values[valuesOffset] = (block3 >> 18) & 16383
		valuesOffset++
		values[valuesOffset] = (block3 >> 4) & 16383
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 15) << 10) | (block4 >> 54)
		valuesOffset++
		values[valuesOffset] = (block4 >> 40) & 16383
		valuesOffset++
		values[valuesOffset] = (block4 >> 26) & 16383
		valuesOffset++
		values[valuesOffset] = (block4 >> 12) & 16383
		valuesOffset++
		block5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block4 & 4095) << 2) | (block5 >> 62)
		valuesOffset++
		values[valuesOffset] = (block5 >> 48) & 16383
		valuesOffset++
		values[valuesOffset] = (block5 >> 34) & 16383
		valuesOffset++
		values[valuesOffset] = (block5 >> 20) & 16383
		valuesOffset++
		values[valuesOffset] = (block5 >> 6) & 16383
		valuesOffset++
		block6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block5 & 63) << 8) | (block6 >> 56)
		valuesOffset++
		values[valuesOffset] = (block6 >> 42) & 16383
		valuesOffset++
		values[valuesOffset] = (block6 >> 28) & 16383
		valuesOffset++
		values[valuesOffset] = (block6 >> 14) & 16383
		valuesOffset++
		values[valuesOffset] = block6 & 16383
		valuesOffset++
	}
}

func (b *BulkOperationPacked14) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 6) | (byte1 >> 2)
		valuesOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte3 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte1 & 3) << 12) | (byte2 << 4) | (byte3 >> 4)
		valuesOffset++
		byte4 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte5 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte3 & 15) << 10) | (byte4 << 2) | (byte5 >> 6)
		valuesOffset++
		byte6 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte5 & 63) << 8) | byte6
		valuesOffset++
	}
}

//func (b *BulkOperationPacked14) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
//	blocksOffset, valuesOffset := 0, 0
//	for i := 0; i < iterations; i++ {
//		byte0 := int32(blocks[blocksOffset])
//		blocksOffset++
//		byte1 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = (byte0 << 6) | (byte1 >> 2)
//		valuesOffset++
//		byte2 := int32(blocks[blocksOffset])
//		blocksOffset++
//		byte3 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = ((byte1 & 3) << 12) | (byte2 << 4) | (byte3 >> 4)
//		valuesOffset++
//		byte4 := int32(blocks[blocksOffset])
//		blocksOffset++
//		byte5 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = ((byte3 & 15) << 10) | (byte4 << 2) | (byte5 >> 6)
//		valuesOffset++
//		byte6 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = ((byte5 & 63) << 8) | byte6
//		valuesOffset++
//	}
//}
