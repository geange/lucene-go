package bulkoperation

type Packed12 struct {
	*BulkOperationPacked
}

func NewPacked12() *Packed12 {
	return &Packed12{NewPacked(12)}
}

func (b *Packed12) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = block0 >> 52
		valuesOffset++
		values[valuesOffset] = (block0 >> 40) & 4095
		valuesOffset++
		values[valuesOffset] = (block0 >> 28) & 4095
		valuesOffset++
		values[valuesOffset] = (block0 >> 16) & 4095
		valuesOffset++
		values[valuesOffset] = (block0 >> 4) & 4095
		valuesOffset++
		block1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block0 & 15) << 8) | (block1 >> 56)
		valuesOffset++
		values[valuesOffset] = (block1 >> 44) & 4095
		valuesOffset++
		values[valuesOffset] = (block1 >> 32) & 4095
		valuesOffset++
		values[valuesOffset] = (block1 >> 20) & 4095
		valuesOffset++
		values[valuesOffset] = (block1 >> 8) & 4095
		valuesOffset++
		block2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = ((block1 & 255) << 4) | (block2 >> 60)
		valuesOffset++
		values[valuesOffset] = (block2 >> 48) & 4095
		valuesOffset++
		values[valuesOffset] = (block2 >> 36) & 4095
		valuesOffset++
		values[valuesOffset] = (block2 >> 24) & 4095
		valuesOffset++
		values[valuesOffset] = (block2 >> 12) & 4095
		valuesOffset++
		values[valuesOffset] = block2 & 4095
		valuesOffset++
	}
}

func (b *Packed12) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		byte0 := uint64(blocks[blocksOffset])
		blocksOffset++
		byte1 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (byte0 << 4) | (byte1 >> 4)
		valuesOffset++
		byte2 := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = ((byte1 & 15) << 8) | byte2
		valuesOffset++
	}
}
