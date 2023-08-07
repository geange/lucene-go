package packed

type BulkOperationPacked15 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked15() *BulkOperationPacked15 {
	return &BulkOperationPacked15{NewBulkOperationPacked(15)}
}

func (b *BulkOperationPacked15) DecodeLongToLong(blocks, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 49
		valuesOffset++
		values[valuesOffset] = (block0 >> 34) & 32767
		valuesOffset++
		values[valuesOffset] = (block0 >> 19) & 32767
		valuesOffset++
		values[valuesOffset] = (block0 >> 4) & 32767
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 15) << 11) | (block1 >> 53)
		valuesOffset++
		values[valuesOffset] = (block1 >> 38) & 32767
		valuesOffset++
		values[valuesOffset] = (block1 >> 23) & 32767
		valuesOffset++
		values[valuesOffset] = (block1 >> 8) & 32767
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 255) << 7) | (block2 >> 57)
		valuesOffset++
		values[valuesOffset] = (block2 >> 42) & 32767
		valuesOffset++
		values[valuesOffset] = (block2 >> 27) & 32767
		valuesOffset++
		values[valuesOffset] = (block2 >> 12) & 32767
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 4095) << 3) | (block3 >> 61)
		valuesOffset++
		values[valuesOffset] = (block3 >> 46) & 32767
		valuesOffset++
		values[valuesOffset] = (block3 >> 31) & 32767
		valuesOffset++
		values[valuesOffset] = (block3 >> 16) & 32767
		valuesOffset++
		values[valuesOffset] = (block3 >> 1) & 32767
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 1) << 14) | (block4 >> 50)
		valuesOffset++
		values[valuesOffset] = (block4 >> 35) & 32767
		valuesOffset++
		values[valuesOffset] = (block4 >> 20) & 32767
		valuesOffset++
		values[valuesOffset] = (block4 >> 5) & 32767
		valuesOffset++
		block5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block4 & 31) << 10) | (block5 >> 54)
		valuesOffset++
		values[valuesOffset] = (block5 >> 39) & 32767
		valuesOffset++
		values[valuesOffset] = (block5 >> 24) & 32767
		valuesOffset++
		values[valuesOffset] = (block5 >> 9) & 32767
		valuesOffset++
		block6 := blocks[blocksOffset]
		values[valuesOffset] = ((block5 & 511) << 6) | (block6 >> 58)
		valuesOffset++
		values[valuesOffset] = (block6 >> 43) & 32767
		valuesOffset++
		values[valuesOffset] = (block6 >> 28) & 32767
		valuesOffset++
		values[valuesOffset] = (block6 >> 13) & 32767
		valuesOffset++
		block7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block6 & 8191) << 2) | (block7 >> 62)
		valuesOffset++
		values[valuesOffset] = (block7 >> 47) & 32767
		valuesOffset++
		values[valuesOffset] = (block7 >> 32) & 32767
		valuesOffset++
		values[valuesOffset] = (block7 >> 17) & 32767
		valuesOffset++
		values[valuesOffset] = (block7 >> 2) & 32767
		valuesOffset++
		block8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block7 & 3) << 13) | (block8 >> 51)
		valuesOffset++
		values[valuesOffset] = (block8 >> 36) & 32767
		valuesOffset++
		values[valuesOffset] = (block8 >> 21) & 32767
		valuesOffset++
		values[valuesOffset] = (block8 >> 6) & 32767
		valuesOffset++
		block9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block8 & 63) << 9) | (block9 >> 55)
		valuesOffset++
		values[valuesOffset] = (block9 >> 40) & 32767
		valuesOffset++
		values[valuesOffset] = (block9 >> 25) & 32767
		valuesOffset++
		values[valuesOffset] = (block9 >> 10) & 32767
		valuesOffset++
		block10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block9 & 1023) << 5) | (block10 >> 59)
		valuesOffset++
		values[valuesOffset] = (block10 >> 44) & 32767
		valuesOffset++
		values[valuesOffset] = (block10 >> 29) & 32767
		valuesOffset++
		values[valuesOffset] = (block10 >> 14) & 32767
		valuesOffset++
		block11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block10 & 16383) << 1) | (block11 >> 63)
		valuesOffset++
		values[valuesOffset] = (block11 >> 48) & 32767
		valuesOffset++
		values[valuesOffset] = (block11 >> 33) & 32767
		valuesOffset++
		values[valuesOffset] = (block11 >> 18) & 32767
		valuesOffset++
		values[valuesOffset] = (block11 >> 3) & 32767
		valuesOffset++
		block12 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block11 & 7) << 12) | (block12 >> 52)
		valuesOffset++
		values[valuesOffset] = (block12 >> 37) & 32767
		valuesOffset++
		values[valuesOffset] = (block12 >> 22) & 32767
		valuesOffset++
		values[valuesOffset] = (block12 >> 7) & 32767
		valuesOffset++
		block13 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block12 & 127) << 8) | (block13 >> 56)
		valuesOffset++
		values[valuesOffset] = (block13 >> 41) & 32767
		valuesOffset++
		values[valuesOffset] = (block13 >> 26) & 32767
		valuesOffset++
		values[valuesOffset] = (block13 >> 11) & 32767
		valuesOffset++
		block14 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block13 & 2047) << 4) | (block14 >> 60)
		valuesOffset++
		values[valuesOffset] = (block14 >> 45) & 32767
		valuesOffset++
		values[valuesOffset] = (block14 >> 30) & 32767
		valuesOffset++
		values[valuesOffset] = (block14 >> 15) & 32767
		valuesOffset++
		values[valuesOffset] = block14 & 32767
		valuesOffset++
	}
}

func (b *BulkOperationPacked15) DecodeByteToLong(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 7) | (byte1 >> 1)
		valuesOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte3 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte1 & 1) << 14) | (byte2 << 6) | (byte3 >> 2)
		valuesOffset++
		byte4 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte5 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte3 & 3) << 13) | (byte4 << 5) | (byte5 >> 3)
		valuesOffset++
		byte6 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte7 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte5 & 7) << 12) | (byte6 << 4) | (byte7 >> 4)
		valuesOffset++
		byte8 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte9 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte7 & 15) << 11) | (byte8 << 3) | (byte9 >> 5)
		valuesOffset++
		byte10 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte11 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte9 & 31) << 10) | (byte10 << 2) | (byte11 >> 6)
		valuesOffset++
		byte12 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte13 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte11 & 63) << 9) | (byte12 << 1) | (byte13 >> 7)
		valuesOffset++
		byte14 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte13 & 127) << 8) | byte14
		valuesOffset++
	}
}

func (b *BulkOperationPacked15) DecodeByteToInt(blocks []byte, values []uint32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 7) | (byte1 >> 1)
		valuesOffset++
		byte2 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte3 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte1 & 1) << 14) | (byte2 << 6) | (byte3 >> 2)
		valuesOffset++
		byte4 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte5 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte3 & 3) << 13) | (byte4 << 5) | (byte5 >> 3)
		valuesOffset++
		byte6 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte7 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte5 & 7) << 12) | (byte6 << 4) | (byte7 >> 4)
		valuesOffset++
		byte8 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte9 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte7 & 15) << 11) | (byte8 << 3) | (byte9 >> 5)
		valuesOffset++
		byte10 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte11 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte9 & 31) << 10) | (byte10 << 2) | (byte11 >> 6)
		valuesOffset++
		byte12 := uint32(blocks[blocksOffset])
		blocksOffset++
		byte13 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte11 & 63) << 9) | (byte12 << 1) | (byte13 >> 7)
		valuesOffset++
		byte14 := uint32(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte13 & 127) << 8) | byte14
		valuesOffset++
	}
}
