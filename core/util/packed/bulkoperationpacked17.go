package packed

type BulkOperationPacked17 struct {
	*BulkOperationPacked
}

func NewBulkOperationPacked17() *BulkOperationPacked17 {
	return &BulkOperationPacked17{NewBulkOperationPacked(17)}
}

func (b *BulkOperationPacked17) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 47
		valuesOffset++
		values[valuesOffset] = (block0 >> 30) & 131071
		valuesOffset++
		values[valuesOffset] = (block0 >> 13) & 131071
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 8191) << 4) | (block1 >> 60)
		valuesOffset++
		values[valuesOffset] = (block1 >> 43) & 131071
		valuesOffset++
		values[valuesOffset] = (block1 >> 26) & 131071
		valuesOffset++
		values[valuesOffset] = (block1 >> 9) & 131071
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 511) << 8) | (block2 >> 56)
		valuesOffset++
		values[valuesOffset] = (block2 >> 39) & 131071
		valuesOffset++
		values[valuesOffset] = (block2 >> 22) & 131071
		valuesOffset++
		values[valuesOffset] = (block2 >> 5) & 131071
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 31) << 12) | (block3 >> 52)
		valuesOffset++
		values[valuesOffset] = (block3 >> 35) & 131071
		valuesOffset++
		values[valuesOffset] = (block3 >> 18) & 131071
		valuesOffset++
		values[valuesOffset] = (block3 >> 1) & 131071
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 1) << 16) | (block4 >> 48)
		valuesOffset++
		values[valuesOffset] = (block4 >> 31) & 131071
		valuesOffset++
		values[valuesOffset] = (block4 >> 14) & 131071
		valuesOffset++
		block5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block4 & 16383) << 3) | (block5 >> 61)
		valuesOffset++
		values[valuesOffset] = (block5 >> 44) & 131071
		valuesOffset++
		values[valuesOffset] = (block5 >> 27) & 131071
		valuesOffset++
		values[valuesOffset] = (block5 >> 10) & 131071
		valuesOffset++
		block6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block5 & 1023) << 7) | (block6 >> 57)
		valuesOffset++
		values[valuesOffset] = (block6 >> 40) & 131071
		valuesOffset++
		values[valuesOffset] = (block6 >> 23) & 131071
		valuesOffset++
		values[valuesOffset] = (block6 >> 6) & 131071
		valuesOffset++
		block7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block6 & 63) << 11) | (block7 >> 53)
		valuesOffset++
		values[valuesOffset] = (block7 >> 36) & 131071
		valuesOffset++
		values[valuesOffset] = (block7 >> 19) & 131071
		valuesOffset++
		values[valuesOffset] = (block7 >> 2) & 131071
		valuesOffset++
		block8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block7 & 3) << 15) | (block8 >> 49)
		valuesOffset++
		values[valuesOffset] = (block8 >> 32) & 131071
		valuesOffset++
		values[valuesOffset] = (block8 >> 15) & 131071
		valuesOffset++
		block9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block8 & 32767) << 2) | (block9 >> 62)
		valuesOffset++
		values[valuesOffset] = (block9 >> 45) & 131071
		valuesOffset++
		values[valuesOffset] = (block9 >> 28) & 131071
		valuesOffset++
		values[valuesOffset] = (block9 >> 11) & 131071
		valuesOffset++
		block10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block9 & 2047) << 6) | (block10 >> 58)
		valuesOffset++
		values[valuesOffset] = (block10 >> 41) & 131071
		valuesOffset++
		values[valuesOffset] = (block10 >> 24) & 131071
		valuesOffset++
		values[valuesOffset] = (block10 >> 7) & 131071
		valuesOffset++
		block11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block10 & 127) << 10) | (block11 >> 54)
		valuesOffset++
		values[valuesOffset] = (block11 >> 37) & 131071
		valuesOffset++
		values[valuesOffset] = (block11 >> 20) & 131071
		valuesOffset++
		values[valuesOffset] = (block11 >> 3) & 131071
		valuesOffset++
		block12 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block11 & 7) << 14) | (block12 >> 50)
		valuesOffset++
		values[valuesOffset] = (block12 >> 33) & 131071
		valuesOffset++
		values[valuesOffset] = (block12 >> 16) & 131071
		valuesOffset++
		block13 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block12 & 65535) << 1) | (block13 >> 63)
		valuesOffset++
		values[valuesOffset] = (block13 >> 46) & 131071
		valuesOffset++
		values[valuesOffset] = (block13 >> 29) & 131071
		valuesOffset++
		values[valuesOffset] = (block13 >> 12) & 131071
		valuesOffset++
		block14 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block13 & 4095) << 5) | (block14 >> 59)
		valuesOffset++
		values[valuesOffset] = (block14 >> 42) & 131071
		valuesOffset++
		values[valuesOffset] = (block14 >> 25) & 131071
		valuesOffset++
		values[valuesOffset] = (block14 >> 8) & 131071
		valuesOffset++
		block15 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block14 & 255) << 9) | (block15 >> 55)
		valuesOffset++
		values[valuesOffset] = (block15 >> 38) & 131071
		valuesOffset++
		values[valuesOffset] = (block15 >> 21) & 131071
		valuesOffset++
		values[valuesOffset] = (block15 >> 4) & 131071
		valuesOffset++
		block16 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block15 & 15) << 13) | (block16 >> 51)
		valuesOffset++
		values[valuesOffset] = (block16 >> 34) & 131071
		valuesOffset++
		values[valuesOffset] = (block16 >> 17) & 131071
		valuesOffset++
		values[valuesOffset] = block16 & 131071
		valuesOffset++
	}
}

func (b *BulkOperationPacked17) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 9) | (byte1 << 1) | (byte2 >> 7)
		valuesOffset++
		byte3 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte4 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte2 & 127) << 10) | (byte3 << 2) | (byte4 >> 6)
		valuesOffset++
		byte5 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte6 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte4 & 63) << 11) | (byte5 << 3) | (byte6 >> 5)
		valuesOffset++
		byte7 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte8 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte6 & 31) << 12) | (byte7 << 4) | (byte8 >> 4)
		valuesOffset++
		byte9 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte10 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte8 & 15) << 13) | (byte9 << 5) | (byte10 >> 3)
		valuesOffset++
		byte11 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte12 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte10 & 7) << 14) | (byte11 << 6) | (byte12 >> 2)
		valuesOffset++
		byte13 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte14 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte12 & 3) << 15) | (byte13 << 7) | (byte14 >> 1)
		valuesOffset++
		byte15 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte16 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte14 & 1) << 16) | (byte15 << 8) | byte16
		valuesOffset++
	}
}
