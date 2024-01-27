package packed

type BulkOperationPacked7 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked7() *BulkOperationPacked7 {
	return &BulkOperationPacked7{NewBulkOperationPacked(7)}
}

func (b *BulkOperationPacked7) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0

	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 57
		valuesOffset++
		values[valuesOffset] = (block0 >> 50) & 127
		valuesOffset++
		values[valuesOffset] = (block0 >> 43) & 127
		valuesOffset++
		values[valuesOffset] = (block0 >> 36) & 127
		valuesOffset++
		values[valuesOffset] = (block0 >> 29) & 127
		valuesOffset++
		values[valuesOffset] = (block0 >> 22) & 127
		valuesOffset++
		values[valuesOffset] = (block0 >> 15) & 127
		valuesOffset++
		values[valuesOffset] = (block0 >> 8) & 127
		valuesOffset++
		values[valuesOffset] = (block0 >> 1) & 127
		valuesOffset++

		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 1) << 6) | (block1 >> 58)
		valuesOffset++
		values[valuesOffset] = (block1 >> 51) & 127
		valuesOffset++
		values[valuesOffset] = (block1 >> 44) & 127
		valuesOffset++
		values[valuesOffset] = (block1 >> 37) & 127
		valuesOffset++
		values[valuesOffset] = (block1 >> 30) & 127
		valuesOffset++
		values[valuesOffset] = (block1 >> 23) & 127
		valuesOffset++
		values[valuesOffset] = (block1 >> 16) & 127
		valuesOffset++
		values[valuesOffset] = (block1 >> 9) & 127
		valuesOffset++
		values[valuesOffset] = (block1 >> 2) & 127
		valuesOffset++

		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 5) | (block2 >> 59)
		valuesOffset++
		values[valuesOffset] = (block2 >> 52) & 127
		valuesOffset++
		values[valuesOffset] = (block2 >> 45) & 127
		valuesOffset++
		values[valuesOffset] = (block2 >> 38) & 127
		valuesOffset++
		values[valuesOffset] = (block2 >> 31) & 127
		valuesOffset++
		values[valuesOffset] = (block2 >> 24) & 127
		valuesOffset++
		values[valuesOffset] = (block2 >> 17) & 127
		valuesOffset++
		values[valuesOffset] = (block2 >> 10) & 127
		valuesOffset++
		values[valuesOffset] = (block2 >> 3) & 127
		valuesOffset++

		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 7) << 4) | (block3 >> 60)
		valuesOffset++
		values[valuesOffset] = (block3 >> 53) & 127
		valuesOffset++
		values[valuesOffset] = (block3 >> 46) & 127
		valuesOffset++
		values[valuesOffset] = (block3 >> 39) & 127
		valuesOffset++
		values[valuesOffset] = (block3 >> 32) & 127
		valuesOffset++
		values[valuesOffset] = (block3 >> 25) & 127
		valuesOffset++
		values[valuesOffset] = (block3 >> 18) & 127
		valuesOffset++
		values[valuesOffset] = (block3 >> 11) & 127
		valuesOffset++
		values[valuesOffset] = (block3 >> 4) & 127
		valuesOffset++

		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 15) << 3) | (block4 >> 61)
		valuesOffset++
		values[valuesOffset] = (block4 >> 54) & 127
		valuesOffset++
		values[valuesOffset] = (block4 >> 47) & 127
		valuesOffset++
		values[valuesOffset] = (block4 >> 40) & 127
		valuesOffset++
		values[valuesOffset] = (block4 >> 33) & 127
		valuesOffset++
		values[valuesOffset] = (block4 >> 26) & 127
		valuesOffset++
		values[valuesOffset] = (block4 >> 19) & 127
		valuesOffset++
		values[valuesOffset] = (block4 >> 12) & 127
		valuesOffset++
		values[valuesOffset] = (block4 >> 5) & 127
		valuesOffset++

		block5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block4 & 31) << 2) | (block5 >> 62)
		valuesOffset++
		values[valuesOffset] = (block5 >> 55) & 127
		valuesOffset++
		values[valuesOffset] = (block5 >> 48) & 127
		valuesOffset++
		values[valuesOffset] = (block5 >> 41) & 127
		valuesOffset++
		values[valuesOffset] = (block5 >> 34) & 127
		valuesOffset++
		values[valuesOffset] = (block5 >> 27) & 127
		valuesOffset++
		values[valuesOffset] = (block5 >> 20) & 127
		valuesOffset++
		values[valuesOffset] = (block5 >> 13) & 127
		valuesOffset++
		values[valuesOffset] = (block5 >> 6) & 127
		valuesOffset++

		block6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block5 & 63) << 1) | (block6 >> 63)
		valuesOffset++
		values[valuesOffset] = (block6 >> 56) & 127
		valuesOffset++
		values[valuesOffset] = (block6 >> 49) & 127
		valuesOffset++
		values[valuesOffset] = (block6 >> 42) & 127
		valuesOffset++
		values[valuesOffset] = (block6 >> 35) & 127
		valuesOffset++
		values[valuesOffset] = (block6 >> 28) & 127
		valuesOffset++
		values[valuesOffset] = (block6 >> 21) & 127
		valuesOffset++
		values[valuesOffset] = (block6 >> 14) & 127
		valuesOffset++
		values[valuesOffset] = (block6 >> 7) & 127
		valuesOffset++
		values[valuesOffset] = block6 & 127
		valuesOffset++
	}
}

func (b *BulkOperationPacked7) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = byte0 >> 1
		valuesOffset++

		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte0 & 1) << 6) | (byte1 >> 2)
		valuesOffset++

		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte1 & 3) << 5) | (byte2 >> 3)
		valuesOffset++

		byte3 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte2 & 7) << 4) | (byte3 >> 4)
		valuesOffset++

		byte4 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte3 & 15) << 3) | (byte4 >> 5)
		valuesOffset++

		byte5 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte4 & 31) << 2) | (byte5 >> 6)
		valuesOffset++

		byte6 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte5 & 63) << 1) | (byte6 >> 7)
		valuesOffset++

		values[valuesOffset] = byte6 & 127
		valuesOffset++
	}
}
