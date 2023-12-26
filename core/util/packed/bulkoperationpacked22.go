package packed

type BulkOperationPacked22 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked22() *BulkOperationPacked22 {
	return &BulkOperationPacked22{NewBulkOperationPacked(22)}
}

func (b *BulkOperationPacked22) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 42
		valuesOffset++
		values[valuesOffset] = (block0 >> 20) & 4194303
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 1048575) << 2) | (block1 >> 62)
		valuesOffset++
		values[valuesOffset] = (block1 >> 40) & 4194303
		valuesOffset++
		values[valuesOffset] = (block1 >> 18) & 4194303
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 262143) << 4) | (block2 >> 60)
		valuesOffset++
		values[valuesOffset] = (block2 >> 38) & 4194303
		valuesOffset++
		values[valuesOffset] = (block2 >> 16) & 4194303
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 65535) << 6) | (block3 >> 58)
		valuesOffset++
		values[valuesOffset] = (block3 >> 36) & 4194303
		valuesOffset++
		values[valuesOffset] = (block3 >> 14) & 4194303
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 16383) << 8) | (block4 >> 56)
		valuesOffset++
		values[valuesOffset] = (block4 >> 34) & 4194303
		valuesOffset++
		values[valuesOffset] = (block4 >> 12) & 4194303
		valuesOffset++
		block5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block4 & 4095) << 10) | (block5 >> 54)
		valuesOffset++
		values[valuesOffset] = (block5 >> 32) & 4194303
		valuesOffset++
		values[valuesOffset] = (block5 >> 10) & 4194303
		valuesOffset++
		block6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block5 & 1023) << 12) | (block6 >> 52)
		valuesOffset++
		values[valuesOffset] = (block6 >> 30) & 4194303
		valuesOffset++
		values[valuesOffset] = (block6 >> 8) & 4194303
		valuesOffset++
		block7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block6 & 255) << 14) | (block7 >> 50)
		valuesOffset++
		values[valuesOffset] = (block7 >> 28) & 4194303
		valuesOffset++
		values[valuesOffset] = (block7 >> 6) & 4194303
		valuesOffset++
		block8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block7 & 63) << 16) | (block8 >> 48)
		valuesOffset++
		values[valuesOffset] = (block8 >> 26) & 4194303
		valuesOffset++
		values[valuesOffset] = (block8 >> 4) & 4194303
		valuesOffset++
		block9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block8 & 15) << 18) | (block9 >> 46)
		valuesOffset++
		values[valuesOffset] = (block9 >> 24) & 4194303
		valuesOffset++
		values[valuesOffset] = (block9 >> 2) & 4194303
		valuesOffset++
		block10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block9 & 3) << 20) | (block10 >> 44)
		valuesOffset++
		values[valuesOffset] = (block10 >> 22) & 4194303
		valuesOffset++
		values[valuesOffset] = block10 & 4194303
		valuesOffset++
	}
}

func (b *BulkOperationPacked22) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 14) | (byte1 << 6) | (byte2 >> 2)
		valuesOffset++
		byte3 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte4 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte5 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte2 & 3) << 20) | (byte3 << 12) | (byte4 << 4) | (byte5 >> 4)
		valuesOffset++
		byte6 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte7 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte8 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte5 & 15) << 18) | (byte6 << 10) | (byte7 << 2) | (byte8 >> 6)
		valuesOffset++
		byte9 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte10 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte8 & 63) << 16) | (byte9 << 8) | byte10
		valuesOffset++
	}
}

//func (b *BulkOperationPacked22) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
//	blocksOffset, valuesOffset := 0, 0
//	for i := 0; i < iterations; i++ {
//		byte0 := int32(blocks[blocksOffset])
//		blocksOffset++
//		byte1 := int32(blocks[blocksOffset])
//		blocksOffset++
//		byte2 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = (byte0 << 14) | (byte1 << 6) | (byte2 >> 2)
//		valuesOffset++
//		byte3 := int32(blocks[blocksOffset])
//		blocksOffset++
//		byte4 := int32(blocks[blocksOffset])
//		blocksOffset++
//		byte5 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = ((byte2 & 3) << 20) | (byte3 << 12) | (byte4 << 4) | (byte5 >> 4)
//		valuesOffset++
//		byte6 := int32(blocks[blocksOffset])
//		blocksOffset++
//		byte7 := int32(blocks[blocksOffset])
//		blocksOffset++
//		byte8 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = ((byte5 & 15) << 18) | (byte6 << 10) | (byte7 << 2) | (byte8 >> 6)
//		valuesOffset++
//		byte9 := int32(blocks[blocksOffset])
//		blocksOffset++
//		byte10 := int32(blocks[blocksOffset])
//		blocksOffset++
//		values[valuesOffset] = ((byte8 & 63) << 16) | (byte9 << 8) | byte10
//		valuesOffset++
//	}
//}
