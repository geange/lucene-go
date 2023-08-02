package bulkoperation

type Packed20 struct {
	*BulkOperationPacked
}

func NewPacked20() *Packed20 {
	return &Packed20{NewPacked(20)}
}

func (b *Packed20) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 44
		valuesOffset++
		values[valuesOffset] = (block0 >> 24) & 1048575
		valuesOffset++
		values[valuesOffset] = (block0 >> 4) & 1048575
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 15) << 16) | (block1 >> 48)
		valuesOffset++
		values[valuesOffset] = (block1 >> 28) & 1048575
		valuesOffset++
		values[valuesOffset] = (block1 >> 8) & 1048575
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 255) << 12) | (block2 >> 52)
		valuesOffset++
		values[valuesOffset] = (block2 >> 32) & 1048575
		valuesOffset++
		values[valuesOffset] = (block2 >> 12) & 1048575
		valuesOffset++
		block3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block2 & 4095) << 8) | (block3 >> 56)
		valuesOffset++
		values[valuesOffset] = (block3 >> 36) & 1048575
		valuesOffset++
		values[valuesOffset] = (block3 >> 16) & 1048575
		valuesOffset++
		block4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block3 & 65535) << 4) | (block4 >> 60)
		valuesOffset++
		values[valuesOffset] = (block4 >> 40) & 1048575
		valuesOffset++
		values[valuesOffset] = (block4 >> 20) & 1048575
		valuesOffset++
		values[valuesOffset] = block4 & 1048575
		valuesOffset++
	}
}

func (b *Packed20) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 12) | (byte1 << 4) | (byte2 >> 4)
		valuesOffset++
		byte3 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte4 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte2 & 15) << 16) | (byte3 << 8) | byte4
		valuesOffset++
	}
}
