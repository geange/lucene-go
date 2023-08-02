package bulkoperation

type Packed2 struct {
	*BulkOperationPacked
}

func NewPacked2() *Packed2 {
	return &Packed2{NewPacked(2)}
}

func (b *Packed2) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := blocks[blocksOffset]
		blocksOffset++
		for shift := 62; shift >= 0; shift -= 2 {
			values[valuesOffset] = uint64((block >> shift) & 3)
			valuesOffset++
		}
	}
}

func (b *Packed2) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := uint64(blocks[blocksOffset])
		blocksOffset++
		values[valuesOffset] = (block >> 6) & 3
		valuesOffset++
		values[valuesOffset] = (block >> 4) & 3
		valuesOffset++
		values[valuesOffset] = (block >> 2) & 3
		valuesOffset++
		values[valuesOffset] = block & 3
		valuesOffset++
	}
}
