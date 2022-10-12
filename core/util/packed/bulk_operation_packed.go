package packed

var (
	_ BulkOperation = &BulkOperationPacked{}
)

// BulkOperationPacked Non-specialized BulkOperation for PackedInts.Format.PACKED.
type BulkOperationPacked struct {
	*bulkOperation

	bitsPerValue   int
	longBlockCount int // 一个block的大小为 [longBlockCount]int64
	longValueCount int // 一个block存储的value的数量为 longValueCount
	byteBlockCount int // 一个block的大小为 [longBlockCount * 8]uint8
	byteValueCount int // 一个block存储的value的数量为 longValueCount
	mask           int64
	intMask        int
}

func NewBulkOperationPacked(bitsPerValue int) *BulkOperationPacked {
	packed := &BulkOperationPacked{bitsPerValue: bitsPerValue}

	// bitsPerValue 值占多少个比特位
	blocks := bitsPerValue
	for blocks&1 == 0 {
		blocks = blocks >> 1
	}
	packed.longBlockCount = blocks
	packed.longValueCount = 64 * packed.longBlockCount / bitsPerValue
	byteBlockCount := 8 * packed.longBlockCount
	byteValueCount := packed.longValueCount
	for (byteBlockCount&1) == 0 && (byteValueCount&1) == 0 {
		byteBlockCount = byteBlockCount >> 1
		byteValueCount = byteValueCount >> 1
	}
	packed.byteBlockCount = byteBlockCount
	packed.byteValueCount = byteValueCount
	if bitsPerValue == 64 {
		packed.mask = ^0
	} else {
		packed.mask = (1 << bitsPerValue) - 1
	}
	packed.intMask = int(packed.mask)
	return packed
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

func (b *BulkOperationPacked) DecodeLongToLong(blocks, values []int64, iterations int) {
	bitsLeft := 64
	valuesOffset, blocksOffset := 0, 0
	for i := 0; i < b.longValueCount*iterations; i++ {
		bitsLeft -= b.bitsPerValue
		if bitsLeft < 0 {
			values[valuesOffset] = ((blocks[blocksOffset] &
				((1 << (b.bitsPerValue + bitsLeft)) - 1)) << -bitsLeft) |
				(blocks[blocksOffset] >> (64 + bitsLeft))
			bitsLeft += 64
			valuesOffset++
			blocksOffset++
		} else {
			values[valuesOffset] = (blocks[blocksOffset] >> bitsLeft) & b.mask
			valuesOffset++
		}
	}
}

func (b *BulkOperationPacked) DecodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	nextValue := 0
	bitsLeft := b.bitsPerValue
	for i := 0; i < iterations*b.byteBlockCount; i++ {
		bytes := blocks[blocksOffset] & 0xFF
		blocksOffset++
		if bitsLeft > 8 {
			// just buffer
			bitsLeft -= 8
			nextValue |= bytes << bitsLeft
		} else {
			// flush
			bits := 8 - bitsLeft
			values[valuesOffset] = nextValue | (bytes >> bits)
			valuesOffset++
			for bits >= b.bitsPerValue {
				bits -= b.bitsPerValue
				values[valuesOffset] = (bytes >> bits) & b.mask
				valuesOffset++
			}
			// then buffer
			bitsLeft = b.bitsPerValue - bits
			nextValue = (bytes & ((1 << bits) - 1)) << bitsLeft
		}
	}
}

func (b *BulkOperationPacked) DecodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	nextValue := 0
	bitsLeft := b.bitsPerValue
	for i := 0; i < iterations*b.byteBlockCount; i++ {
		bytes := blocks[blocksOffset] & 0xFF
		blocksOffset++
		if bitsLeft > 8 {
			// just buffer
			bitsLeft -= 8
			nextValue |= bytes << bitsLeft
		} else {
			// flush
			bits := 8 - bitsLeft
			values[valuesOffset] = nextValue | (bytes >> bits)
			valuesOffset++
			for bits >= b.bitsPerValue {
				bits -= b.bitsPerValue
				values[valuesOffset] = (bytes >> bits) & b.intMask
				valuesOffset++
			}
			// then buffer
			bitsLeft = b.bitsPerValue - bits
			nextValue = (bytes & ((1 << bits) - 1)) << bitsLeft
		}
	}
}

func (b *BulkOperationPacked) EncodeLongToLong(values, blocks []int64, iterations int) {
	valuesOffset, blocksOffset := 0, 0
	nextBlock := int64(0)
	bitsLeft := 64
	for i := 0; i < b.longValueCount*iterations; i++ {
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

func (b *BulkOperationPacked) EncodeLongToBytes(values []int64, blocks []byte, iterations int) {
	valuesOffset, blocksOffset := 0, 0

	nextBlock := 0
	bitsLeft := 8
	for i := 0; i < b.byteValueCount*iterations; i++ {
		v := values[valuesOffset]
		valuesOffset++
		if b.bitsPerValue < bitsLeft {
			// just buffer
			nextBlock |= v << (bitsLeft - b.bitsPerValue)
			bitsLeft -= b.bitsPerValue
		} else {
			// flush as many blocks as possible
			bits := b.bitsPerValue - bitsLeft
			blocks[blocksOffset] = (byte)(nextBlock | (v >> bits))
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

func (b *BulkOperationPacked) EncodeIntToBytes(values []int32, blocks []byte, iterations int) {
	valuesOffset, blocksOffset := 0, 0
	nextBlock := 0
	bitsLeft := 8
	for i := 0; i < b.byteValueCount*iterations; i++ {
		v := values[valuesOffset]
		valuesOffset++

		if b.bitsPerValue < bitsLeft {
			// just buffer
			nextBlock |= v << (bitsLeft - b.bitsPerValue)
			bitsLeft -= b.bitsPerValue
		} else {
			// flush as many blocks as possible
			bits := b.bitsPerValue - bitsLeft
			blocks[blocksOffset] = byte(nextBlock | (v >> bits))
			blocksOffset++
			for bits >= 8 {
				bits -= 8
				blocks[blocksOffset] = byte(v >> bits)
				blocksOffset++
			}
			// then buffer
			bitsLeft = 8 - bits
			nextBlock = (v & ((1 << bits) - 1)) << bitsLeft
		}
	}
}
