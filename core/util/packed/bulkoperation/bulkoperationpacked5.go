package bulkoperation

type Packed5 struct {
	*BulkOperationPacked
}

func NewPacked5() *Packed5 {
	return &Packed5{NewPacked(5)}
}

func (b *Packed5) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 59
		valuesOffset++
		values[valuesOffset] = (block0 >> 54) & 31
		valuesOffset++
		values[valuesOffset] = (block0 >> 49) & 31
		valuesOffset++
		values[valuesOffset] = (block0 >> 44) & 31
		valuesOffset++
		values[valuesOffset] = (block0 >> 39) & 31
		valuesOffset++
		values[valuesOffset] = (block0 >> 34) & 31
		valuesOffset++
		values[valuesOffset] = (block0 >> 29) & 31
		valuesOffset++
		values[valuesOffset] = (block0 >> 24) & 31
		valuesOffset++
		values[valuesOffset] = (block0 >> 19) & 31
		valuesOffset++
		values[valuesOffset] = (block0 >> 14) & 31
		valuesOffset++
		values[valuesOffset] = (block0 >> 9) & 31
		valuesOffset++
		values[valuesOffset] = (block0 >> 4) & 31
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 15) << 1) | (block1 >> 63)
		valuesOffset++
		values[valuesOffset] = (block1 >> 58) & 31
		valuesOffset++
		values[valuesOffset] = (block1 >> 53) & 31
		valuesOffset++
		values[valuesOffset] = (block1 >> 48) & 31
		valuesOffset++
		values[valuesOffset] = (block1 >> 43) & 31
		valuesOffset++
		values[valuesOffset] = (block1 >> 38) & 31
		valuesOffset++
		values[valuesOffset] = (block1 >> 33) & 31
		valuesOffset++
		values[valuesOffset] = (block1 >> 28) & 31
		valuesOffset++
		values[valuesOffset] = (block1 >> 23) & 31
		valuesOffset++
		values[valuesOffset] = (block1 >> 18) & 31
		valuesOffset++
		values[valuesOffset] = (block1 >> 13) & 31
		valuesOffset++
		values[valuesOffset] = (block1 >> 8) & 31
		valuesOffset++
		values[valuesOffset] = (block1 >> 3) & 31
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 7) << 2) | (block2 >> 62)
		valuesOffset++
		values[valuesOffset] = (block2 >> 57) & 31
		valuesOffset++
		values[valuesOffset] = (block2 >> 52) & 31
		valuesOffset++
		values[valuesOffset] = (block2 >> 47) & 31
		valuesOffset++
		values[valuesOffset] = (block2 >> 42) & 31
		valuesOffset++
		values[valuesOffset] = (block2 >> 37) & 31
		valuesOffset++
		values[valuesOffset] = (block2 >> 32) & 31
		valuesOffset++
		values[valuesOffset] = (block2 >> 27) & 31
		valuesOffset++
		values[valuesOffset] = (block2 >> 22) & 31
		valuesOffset++
		values[valuesOffset] = (block2 >> 17) & 31
		valuesOffset++
		values[valuesOffset] = (block2 >> 12) & 31
		valuesOffset++
		values[valuesOffset] = (block2 >> 7) & 31
		valuesOffset++
		values[valuesOffset] = (block2 >> 2) & 31
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 3) << 3) | (block3 >> 61)
		valuesOffset++
		values[valuesOffset] = (block3 >> 56) & 31
		valuesOffset++
		values[valuesOffset] = (block3 >> 51) & 31
		valuesOffset++
		values[valuesOffset] = (block3 >> 46) & 31
		valuesOffset++
		values[valuesOffset] = (block3 >> 41) & 31
		valuesOffset++
		values[valuesOffset] = (block3 >> 36) & 31
		valuesOffset++
		values[valuesOffset] = (block3 >> 31) & 31
		valuesOffset++
		values[valuesOffset] = (block3 >> 26) & 31
		valuesOffset++
		values[valuesOffset] = (block3 >> 21) & 31
		valuesOffset++
		values[valuesOffset] = (block3 >> 16) & 31
		valuesOffset++
		values[valuesOffset] = (block3 >> 11) & 31
		valuesOffset++
		values[valuesOffset] = (block3 >> 6) & 31
		valuesOffset++
		values[valuesOffset] = (block3 >> 1) & 31
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 1) << 4) | (block4 >> 60)
		valuesOffset++
		values[valuesOffset] = (block4 >> 55) & 31
		valuesOffset++
		values[valuesOffset] = (block4 >> 50) & 31
		valuesOffset++
		values[valuesOffset] = (block4 >> 45) & 31
		valuesOffset++
		values[valuesOffset] = (block4 >> 40) & 31
		valuesOffset++
		values[valuesOffset] = (block4 >> 35) & 31
		valuesOffset++
		values[valuesOffset] = (block4 >> 30) & 31
		valuesOffset++
		values[valuesOffset] = (block4 >> 25) & 31
		valuesOffset++
		values[valuesOffset] = (block4 >> 20) & 31
		valuesOffset++
		values[valuesOffset] = (block4 >> 15) & 31
		valuesOffset++
		values[valuesOffset] = (block4 >> 10) & 31
		valuesOffset++
		values[valuesOffset] = (block4 >> 5) & 31
		valuesOffset++
		values[valuesOffset] = block4 & 31
		valuesOffset++
	}
}

func (b *Packed5) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = byte0 >> 3
		valuesOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte0 & 7) << 2) | (byte1 >> 6)
		valuesOffset++
		values[valuesOffset] = (byte1 >> 1) & 31
		valuesOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte1 & 1) << 4) | (byte2 >> 4)
		valuesOffset++
		byte3 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte2 & 15) << 1) | (byte3 >> 7)
		valuesOffset++
		values[valuesOffset] = (byte3 >> 2) & 31
		valuesOffset++
		byte4 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte3 & 3) << 3) | (byte4 >> 5)
		valuesOffset++
		values[valuesOffset] = byte4 & 31
		valuesOffset++
	}
}
