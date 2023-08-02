package bulkoperation

type Packed4 struct {
	*BulkOperationPacked
}

func NewPacked4() *Packed4 {
	return &Packed4{NewPacked(4)}
}

func (b *Packed4) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := blocks[blocksOffset]
		blocksOffset++
		for shift := 60; shift >= 0; shift -= 4 {
			values[valuesOffset] = (block >> shift) & 15
			valuesOffset++
		}
	}
}

func (b *Packed4) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j++ {
		block := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (block >> 4) & 15
		valuesOffset++
		values[valuesOffset] = block & 15
		valuesOffset++
	}
}
