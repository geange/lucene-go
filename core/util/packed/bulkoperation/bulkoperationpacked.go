package bulkoperation

import (
	"math"

	"github.com/geange/lucene-go/core/util/packed/common"
)

var (
	_ common.BulkOperation = &BulkOperationPacked{}
)

type BaseBulkOperation struct {
	Decoder common.Decoder
}

// ComputeIterations
// For every number of bits per value, there is a minimum number of blocks (b) / values (v)
// you need to write in order to reach the next block boundary:
// - 16 bits per value -> b=2, v=1
// - 24 bits per value -> b=3, v=1
// - 50 bits per value -> b=25, v=4
// - 63 bits per value -> b=63, v=8
// - ...
// A bulk read consists in copying iterations*v values that are contained in iterations*b blocks
// into a long[] (higher values of iterations are likely to yield a better throughput):
// this requires n * (b + 8v) bytes of memory. This method computes iterations as ramBudget / (b + 8v)
// (since a long is 8 bytes).
func (b *BaseBulkOperation) ComputeIterations(valueCount, ramBudget int) int {
	iterations := ramBudget / (b.Decoder.ByteBlockCount() + 8*b.Decoder.ByteValueCount())
	if iterations == 0 {
		// at least 1
		return 1
	}

	if (iterations-1)*b.Decoder.ByteValueCount() >= valueCount {
		// don't allocate for more than the size of the reader
		return int(math.Ceil(float64(valueCount) / float64(b.Decoder.ByteValueCount())))
	}
	return iterations
}

func ComputeIterations(decoder common.Decoder, valueCount, ramBudget int) int {
	byteValueCount := decoder.ByteValueCount()

	iterations := ramBudget / (decoder.ByteBlockCount() + 8*byteValueCount)
	if iterations == 0 {
		// at least 1
		return 1
	}

	if (iterations-1)*byteValueCount >= valueCount {
		// don't allocate for more than the size of the reader
		return int(math.Ceil(float64(valueCount) / float64(byteValueCount)))
	}
	return iterations
}

// BulkOperationPacked Non-specialized BulkOperation for PackedInts.Format.PACKED.
type BulkOperationPacked struct {
	decoder        common.Decoder
	bitsPerValue   int // 存放的值占用的比特位
	longBlockCount int // 一个block的大小为 [longBlockCount]uint64
	longValueCount int // 一个block存储的value的数量为 longValueCount
	byteBlockCount int // 一个block的大小为 [longBlockCount * 8]uint8
	byteValueCount int // 一个block存储的value的数量为 longValueCount
	mask           uint64
}

func NewPacked(bitsPerValue int) *BulkOperationPacked {
	pack := &BulkOperationPacked{bitsPerValue: bitsPerValue}

	// bitsPerValue 值占多少个比特位
	blocks := bitsPerValue
	for blocks&1 == 0 {
		blocks = blocks >> 1
	}
	pack.longBlockCount = blocks
	pack.longValueCount = 64 * pack.longBlockCount / bitsPerValue
	byteBlockCount := 8 * pack.longBlockCount
	byteValueCount := pack.longValueCount
	for (byteBlockCount&1) == 0 && (byteValueCount&1) == 0 {
		byteBlockCount = byteBlockCount >> 1
		byteValueCount = byteValueCount >> 1
	}
	pack.byteBlockCount = byteBlockCount
	pack.byteValueCount = byteValueCount
	if bitsPerValue == 64 {
		pack.mask = math.MaxUint64
	} else {
		pack.mask = (1 << bitsPerValue) - 1
	}
	pack.decoder = pack
	return pack
}

func (b *BulkOperationPacked) ComputeIterations(valueCount, ramBudget int) int {
	return ComputeIterations(b.decoder, valueCount, ramBudget)
}

func (b *BulkOperationPacked) LongBlockCount() int {
	return b.byteBlockCount
}

func (b *BulkOperationPacked) LongValueCount() int {
	return b.longValueCount
}

func (b *BulkOperationPacked) ByteBlockCount() int {
	return b.byteBlockCount
}

func (b *BulkOperationPacked) ByteValueCount() int {
	return b.byteValueCount
}

func (b *BulkOperationPacked) DecodeUint64(blocks []uint64, values []uint64, iterations int) {
	bitsLeft := 64
	valuesOffset, blocksOffset := 0, 0
	for i := 0; i < b.longValueCount*iterations; i++ {
		bitsLeft -= b.bitsPerValue
		if bitsLeft < 0 {
			values[valuesOffset] =
				((blocks[blocksOffset] & ((1 << (b.bitsPerValue + bitsLeft)) - 1)) << -bitsLeft) |
					(blocks[blocksOffset+1] >> (64 + bitsLeft))

			valuesOffset++
			blocksOffset++
			bitsLeft += 64
		} else {
			values[valuesOffset] = (blocks[blocksOffset] >> bitsLeft) & b.mask
			valuesOffset++
		}
	}
}

func (b *BulkOperationPacked) DecodeBytes(blocks []byte, values []uint64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	nextValue := uint64(0)
	bitsLeft := b.bitsPerValue
	for i := 0; i < iterations*b.byteBlockCount; i++ {
		bytes := blocks[blocksOffset]
		blocksOffset++
		if bitsLeft > 8 {
			// just buffer
			bitsLeft -= 8
			nextValue |= uint64(bytes) << bitsLeft
		} else {
			// flush
			bits := 8 - bitsLeft
			values[valuesOffset] = nextValue | uint64(bytes>>bits)
			valuesOffset++
			for bits >= b.bitsPerValue {
				bits -= b.bitsPerValue
				values[valuesOffset] = uint64(bytes>>bits) & b.mask
				valuesOffset++
			}
			// then buffer
			bitsLeft = b.bitsPerValue - bits
			nextValue = uint64(bytes&((1<<bits)-1)) << bitsLeft
		}
	}
}

func (b *BulkOperationPacked) EncodeUint64(values []uint64, blocks []uint64, iterations int) {
	valuesOffset, blocksOffset := 0, 0
	nextBlock := uint64(0)
	bitsLeft := 64
	count := b.longValueCount * iterations
	for i := 0; i < count; i++ {
		bitsLeft -= b.bitsPerValue
		if bitsLeft > 0 {
			nextBlock |= values[valuesOffset] << bitsLeft
			valuesOffset++
		} else if bitsLeft == 0 {
			nextBlock |= values[valuesOffset]
			valuesOffset++
			blocks[blocksOffset] = nextBlock
			blocksOffset++
			nextBlock = 0
			bitsLeft = 64
		} else { // bitsLeft < 0
			nextBlock |= values[valuesOffset] >> -bitsLeft
			blocks[blocksOffset] = nextBlock
			blocksOffset++
			nextBlock = (values[valuesOffset] & ((1 << -bitsLeft) - 1)) << (64 + bitsLeft)
			valuesOffset++
			bitsLeft += 64
		}
	}
}

// EncodeBytes
func (b *BulkOperationPacked) EncodeBytes(values []uint64, blocks []byte, iterations int) {
	valuesOffset, blocksOffset := 0, 0

	nextBlock := 0
	bitsLeft := 8

	size := b.byteValueCount * iterations

	for i := 0; i < size; i++ {
		v := values[valuesOffset]
		valuesOffset++
		if b.bitsPerValue < bitsLeft {
			// just buffer
			nextBlock |= int(v << (bitsLeft - b.bitsPerValue))
			bitsLeft -= b.bitsPerValue
		} else {
			// flush as many blocks as possible
			bits := b.bitsPerValue - bitsLeft
			blocks[blocksOffset] = (byte)(nextBlock | int(v>>bits))
			blocksOffset++
			for bits >= 8 {
				bits -= 8
				blocks[blocksOffset] = byte(v >> bits)
				blocksOffset++
			}
			// then buffer
			bitsLeft = 8 - bits
			nextBlock = (int)((v & ((1 << bits) - 1)) << bitsLeft)
		}
	}
}
