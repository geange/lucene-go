package bulkoperation

type Packed13 struct {
	*BulkOperationPacked
}

func NewPacked13() *Packed13 {
	return &Packed13{NewPacked(13)}
}

func (b *Packed13) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 51
		valuesOffset++
		values[valuesOffset] = (block0 >> 38) & 8191
		valuesOffset++
		values[valuesOffset] = (block0 >> 25) & 8191
		valuesOffset++
		values[valuesOffset] = (block0 >> 12) & 8191
		valuesOffset++

		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 4095) << 1) | (block1 >> 63)
		valuesOffset++
		values[valuesOffset] = (block1 >> 50) & 8191
		valuesOffset++
		values[valuesOffset] = (block1 >> 37) & 8191
		valuesOffset++
		values[valuesOffset] = (block1 >> 24) & 8191
		valuesOffset++
		values[valuesOffset] = (block1 >> 11) & 8191
		valuesOffset++

		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 2047) << 2) | (block2 >> 62)
		valuesOffset++
		values[valuesOffset] = (block2 >> 49) & 8191
		valuesOffset++
		values[valuesOffset] = (block2 >> 36) & 8191
		valuesOffset++
		values[valuesOffset] = (block2 >> 23) & 8191
		valuesOffset++
		values[valuesOffset] = (block2 >> 10) & 8191
		valuesOffset++

		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 1023) << 3) | (block3 >> 61)
		valuesOffset++
		values[valuesOffset] = (block3 >> 48) & 8191
		valuesOffset++
		values[valuesOffset] = (block3 >> 35) & 8191
		valuesOffset++
		values[valuesOffset] = (block3 >> 22) & 8191
		valuesOffset++
		values[valuesOffset] = (block3 >> 9) & 8191
		valuesOffset++

		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 511) << 4) | (block4 >> 60)
		valuesOffset++
		values[valuesOffset] = (block4 >> 47) & 8191
		valuesOffset++
		values[valuesOffset] = (block4 >> 34) & 8191
		valuesOffset++
		values[valuesOffset] = (block4 >> 21) & 8191
		valuesOffset++
		values[valuesOffset] = (block4 >> 8) & 8191
		valuesOffset++

		block5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block4 & 255) << 5) | (block5 >> 59)
		valuesOffset++
		values[valuesOffset] = (block5 >> 46) & 8191
		valuesOffset++
		values[valuesOffset] = (block5 >> 33) & 8191
		valuesOffset++
		values[valuesOffset] = (block5 >> 20) & 8191
		valuesOffset++
		values[valuesOffset] = (block5 >> 7) & 8191
		valuesOffset++

		block6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block5 & 127) << 6) | (block6 >> 58)
		valuesOffset++
		values[valuesOffset] = (block6 >> 45) & 8191
		valuesOffset++
		values[valuesOffset] = (block6 >> 32) & 8191
		valuesOffset++
		values[valuesOffset] = (block6 >> 19) & 8191
		valuesOffset++
		values[valuesOffset] = (block6 >> 6) & 8191
		valuesOffset++

		block7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block6 & 63) << 7) | (block7 >> 57)
		valuesOffset++
		values[valuesOffset] = (block7 >> 44) & 8191
		valuesOffset++
		values[valuesOffset] = (block7 >> 31) & 8191
		valuesOffset++
		values[valuesOffset] = (block7 >> 18) & 8191
		valuesOffset++
		values[valuesOffset] = (block7 >> 5) & 8191
		valuesOffset++

		block8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block7 & 31) << 8) | (block8 >> 56)
		valuesOffset++
		values[valuesOffset] = (block8 >> 43) & 8191
		valuesOffset++
		values[valuesOffset] = (block8 >> 30) & 8191
		valuesOffset++
		values[valuesOffset] = (block8 >> 17) & 8191
		valuesOffset++
		values[valuesOffset] = (block8 >> 4) & 8191
		valuesOffset++

		block9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block8 & 15) << 9) | (block9 >> 55)
		valuesOffset++
		values[valuesOffset] = (block9 >> 42) & 8191
		valuesOffset++
		values[valuesOffset] = (block9 >> 29) & 8191
		valuesOffset++
		values[valuesOffset] = (block9 >> 16) & 8191
		valuesOffset++
		values[valuesOffset] = (block9 >> 3) & 8191
		valuesOffset++

		block10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block9 & 7) << 10) | (block10 >> 54)
		valuesOffset++
		values[valuesOffset] = (block10 >> 41) & 8191
		valuesOffset++
		values[valuesOffset] = (block10 >> 28) & 8191
		valuesOffset++
		values[valuesOffset] = (block10 >> 15) & 8191
		valuesOffset++
		values[valuesOffset] = (block10 >> 2) & 8191
		valuesOffset++

		block11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block10 & 3) << 11) | (block11 >> 53)
		valuesOffset++
		values[valuesOffset] = (block11 >> 40) & 8191
		valuesOffset++
		values[valuesOffset] = (block11 >> 27) & 8191
		valuesOffset++
		values[valuesOffset] = (block11 >> 14) & 8191
		valuesOffset++
		values[valuesOffset] = (block11 >> 1) & 8191
		valuesOffset++

		block12 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block11 & 1) << 12) | (block12 >> 52)
		valuesOffset++
		values[valuesOffset] = (block12 >> 39) & 8191
		valuesOffset++
		values[valuesOffset] = (block12 >> 26) & 8191
		valuesOffset++
		values[valuesOffset] = (block12 >> 13) & 8191
		valuesOffset++
		values[valuesOffset] = block12 & 8191
		valuesOffset++
	}
}

func (b *Packed13) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 5) | (byte1 >> 3)
		valuesOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte3 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte1 & 7) << 10) | (byte2 << 2) | (byte3 >> 6)
		valuesOffset++
		byte4 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte3 & 63) << 7) | (byte4 >> 1)
		valuesOffset++
		byte5 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte6 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte4 & 1) << 12) | (byte5 << 4) | (byte6 >> 4)
		valuesOffset++
		byte7 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte8 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte6 & 15) << 9) | (byte7 << 1) | (byte8 >> 7)
		valuesOffset++
		byte9 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte8 & 127) << 6) | (byte9 >> 2)
		valuesOffset++
		byte10 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte11 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte9 & 3) << 11) | (byte10 << 3) | (byte11 >> 5)
		valuesOffset++
		byte12 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte11 & 31) << 8) | byte12
		valuesOffset++
	}
}
