package bulkoperation

import (
	"github.com/geange/lucene-go/core/util/packed/common"
)

var _ common.BulkOperation = &PackedSingleBlock{}

const (
	BLOCK_COUNT = 1
)

type PackedSingleBlock struct {
	bitsPerValue int
	valueCount   int
	mask         uint64
}

func NewPackedSingleBlock(bitsPerValue int) *PackedSingleBlock {
	return &PackedSingleBlock{
		bitsPerValue: bitsPerValue,
		valueCount:   64 / bitsPerValue,
		mask:         1<<bitsPerValue - 1,
	}
}

func (b *PackedSingleBlock) LongBlockCount() int {
	return BLOCK_COUNT
}

func (b *PackedSingleBlock) LongValueCount() int {
	return b.valueCount
}

func (b *PackedSingleBlock) ByteBlockCount() int {
	return 8 * BLOCK_COUNT
}

func (b *PackedSingleBlock) ByteValueCount() int {
	return b.valueCount
}

func (b *PackedSingleBlock) decodeLong(block uint64, values []uint64, valuesOffset int) int {
	values[valuesOffset] = block & b.mask
	valuesOffset++
	for i := 0; i < b.valueCount; i++ {
		block = block >> b.bitsPerValue
		values[valuesOffset] = block & b.mask
		valuesOffset++
	}
	return valuesOffset
}

func (b *PackedSingleBlock) decodeInt(block uint64, values []int32, valuesOffset int) int {
	values[valuesOffset] = int32(block & b.mask)
	valuesOffset++
	for i := 0; i < b.valueCount; i++ {
		block = block >> b.bitsPerValue
		values[valuesOffset] = int32(block & b.mask)
		valuesOffset++
	}
	return valuesOffset
}

func (b *PackedSingleBlock) encodeLong(values []uint64, valuesOffset int) uint64 {
	block := values[valuesOffset]
	valuesOffset++
	for j := 1; j < b.valueCount; j++ {
		block |= values[valuesOffset] << (j * b.bitsPerValue)
		valuesOffset++
	}
	return block
}

func (b *PackedSingleBlock) encodeInt(values []int32, valuesOffset int) uint64 {
	block := uint64(values[valuesOffset]) & 0xFFFFFFFF
	valuesOffset++
	for j := 1; j < b.valueCount; j++ {
		block |= uint64(values[valuesOffset]) & 0xFFFFFFFF << (j * b.bitsPerValue)
		valuesOffset++
	}
	return block
}

func (b *PackedSingleBlock) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := blocks[blocksOffset]
		blocksOffset++
		valuesOffset = b.decodeLong(block, values, valuesOffset)
	}
}

func (b *PackedSingleBlock) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := int64(blocks[blocksOffset])
		blocksOffset++
		valuesOffset = b.decodeLong(uint64(block), values, valuesOffset)
	}
}

func (b *PackedSingleBlock) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
	if b.bitsPerValue > 32 {
		panic("cannot decode")
	}
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := int64(blocks[blocksOffset])
		blocksOffset++
		valuesOffset = b.decodeInt(uint64(block), values, valuesOffset)
	}
}

func (b *PackedSingleBlock) EncodeUint64(values []uint64, blocks []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0

	for i := 0; i < iterations; i++ {
		blocks[blocksOffset] = b.encodeLong(values, valuesOffset)
		blocksOffset++
		valuesOffset += b.valueCount
	}
}

func (b *PackedSingleBlock) EncodeBytes(values []uint64, blocks []byte, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := b.encodeLong(values, valuesOffset)
		valuesOffset += b.valueCount
		blocksOffset = writeLong(block, blocks, blocksOffset)
	}
}

func (b *PackedSingleBlock) EncodeI32ToBytes(values []int32, blocks []byte, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := b.encodeInt(values, valuesOffset)
		valuesOffset += b.valueCount
		blocksOffset = writeLong(block, blocks, blocksOffset)
	}
}

func writeLong(block uint64, blocks []byte, blocksOffset int) int {
	for j := 1; j <= 8; j++ {
		blocks[blocksOffset] = byte(block >> (64 - (j << 3)))
		blocksOffset++
	}
	return blocksOffset
}
