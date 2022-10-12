package packed

var _ BulkOperation = &BulkOperationPackedSingleBlock{}

const (
	BLOCK_COUNT = 1
)

type BulkOperationPackedSingleBlock struct {
	bitsPerValue int
	valueCount   int
	mask         int64
}

func NewBulkOperationPackedSingleBlock(bitsPerValue int) *BulkOperationPackedSingleBlock {
	return &BulkOperationPackedSingleBlock{
		bitsPerValue: bitsPerValue,
		valueCount:   64 / bitsPerValue,
		mask:         1<<bitsPerValue - 1,
	}
}

func (b *BulkOperationPackedSingleBlock) LongBlockCount() int {
	return BLOCK_COUNT
}

func (b *BulkOperationPackedSingleBlock) LongValueCount() int {
	return b.valueCount
}

func (b *BulkOperationPackedSingleBlock) ByteBlockCount() int {
	return 8 * BLOCK_COUNT
}

func (b *BulkOperationPackedSingleBlock) ByteValueCount() int {
	return b.valueCount
}

func readLong(blocks []byte, blocksOffset int) int64 {
	return int64(blocks[blocksOffset]&0xFF)<<56 |
		int64(blocks[blocksOffset+1]&0xFF)<<48 |
		int64(blocks[blocksOffset+2]&0xFF)<<40 |
		int64(blocks[blocksOffset+3]&0xFF)<<32 |
		int64(blocks[blocksOffset+4]&0xFF)<<24 |
		int64(blocks[blocksOffset+5]&0xFF)<<16 |
		int64(blocks[blocksOffset+6]&0xFF)<<8 |
		int64(blocks[blocksOffset+7])&0xFF
}

func (b *BulkOperationPackedSingleBlock) decodeLong(block int64, values []int64, valuesOffset int) int {
	values[valuesOffset] = block & b.mask
	valuesOffset++
	for i := 0; i < b.valueCount; i++ {
		block = block >> b.mask
		values[valuesOffset] = block & b.mask
		valuesOffset++
	}
	return valuesOffset
}

func (b *BulkOperationPackedSingleBlock) decodeInt(block int64, values []int32, valuesOffset int) int {
	values[valuesOffset] = int32(block & b.mask)
	valuesOffset++
	for i := 0; i < b.valueCount; i++ {
		block = block >> b.mask
		values[valuesOffset] = int32(block & b.mask)
		valuesOffset++
	}
	return valuesOffset
}

func (b *BulkOperationPackedSingleBlock) encodeLong(values []int64, valuesOffset int) int64 {
	block := values[valuesOffset]
	valuesOffset++
	for j := 1; j < b.valueCount; j++ {
		block |= values[valuesOffset] << (j * b.bitsPerValue)
		valuesOffset++
	}
	return block
}

func (b *BulkOperationPackedSingleBlock) encodeInt(values []int32, valuesOffset int) int64 {
	block := int64(values[valuesOffset]) & 0xFFFFFFFF
	valuesOffset++
	for j := 1; j < b.valueCount; j++ {
		block |= (int64(values[valuesOffset]) & 0xFFFFFFFF) << (j * b.bitsPerValue)
		valuesOffset++
	}
	return block
}

func (b *BulkOperationPackedSingleBlock) DecodeLongToLong(blocks, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := blocks[blocksOffset]
		blocksOffset++
		valuesOffset = b.decodeLong(block, values, valuesOffset)
	}
}

func (b *BulkOperationPackedSingleBlock) DecodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := int64(blocks[blocksOffset])
		blocksOffset++
		valuesOffset = b.decodeLong(block, values, valuesOffset)
	}
}

func (b *BulkOperationPackedSingleBlock) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
	if b.bitsPerValue > 32 {
		panic("cannot decode")
	}
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := int64(blocks[blocksOffset])
		blocksOffset++
		valuesOffset = b.decodeInt(block, values, valuesOffset)
	}
}

func (b *BulkOperationPackedSingleBlock) EncodeLongToLong(values, blocks []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0

	for i := 0; i < iterations; i++ {
		blocks[blocksOffset] = b.encodeLong(values, valuesOffset)
		blocksOffset++
		valuesOffset += b.valueCount
	}
}

func (b *BulkOperationPackedSingleBlock) EncodeLongToBytes(values []int64, blocks []byte, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := b.encodeLong(values, valuesOffset)
		valuesOffset += b.valueCount
		blocksOffset = writeLong(block, blocks, blocksOffset)
	}
}

func (b *BulkOperationPackedSingleBlock) EncodeIntToBytes(values []int32, blocks []byte, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := b.encodeInt(values, valuesOffset)
		valuesOffset += b.valueCount
		blocksOffset = writeLong(block, blocks, blocksOffset)
	}
}
