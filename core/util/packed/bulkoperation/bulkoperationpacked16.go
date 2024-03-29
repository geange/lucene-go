package bulkoperation

type Packed16 struct {
	*BulkOperationPacked
}

func NewPacked16() *Packed16 {
	return &Packed16{NewPacked(16)}
}

func (b *Packed16) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := blocks[blocksOffset]
		blocksOffset++
		for shift := 48; shift >= 0; shift -= 16 {
			values[valuesOffset] = (block >> shift) & 65535
			valuesOffset++
		}
	}
}

func (b *Packed16) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j++ {
		block0 := uint64(blocks[blocksOffset]) << 8
		block1 := uint64(blocks[blocksOffset+1])
		values[valuesOffset] = block0 | block1
		valuesOffset++
		blocksOffset += 2
	}
}
