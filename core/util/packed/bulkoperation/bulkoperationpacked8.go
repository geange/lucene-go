package bulkoperation

type Packed8 struct {
	*BulkOperationPacked
}

func NewPacked8() *Packed8 {
	return &Packed8{NewPacked(8)}
}

func (b *Packed8) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := blocks[blocksOffset]
		blocksOffset++
		for shift := 56; shift >= 0; shift -= 8 {
			values[valuesOffset] = (block >> shift) & 255
			valuesOffset++
		}
	}
}

func (b *Packed8) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j++ {
		values[valuesOffset] = uint64(blocks[blocksOffset])
		blocksOffset++
		valuesOffset++
	}
}
