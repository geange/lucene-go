package packed

import (
	"context"
	"github.com/geange/lucene-go/core/store"
)

const (
	Packed64BlockSize = 64                    // 32 = int, 64 = long
	Packed64BlockBits = 6                     // The #bits representing BLOCK_SIZE
	Packed64ModMask   = Packed64BlockSize - 1 // x % BLOCK_SIZE
)

var _ Mutable = &Packed64{}

type Packed64 struct {
	*MutableImpl

	// Values are stores contiguously in the blocks array.
	blocks []uint64

	// A right-aligned mask of width BitsPerValue used by get(int).
	maskRight uint64

	// Optimization: Saves one lookup in get(int).
	bpvMinusBlockSize int
}

// NewPacked64 Creates an array with the internal structures adjusted for the given limits and initialized to 0.
// Params: 	valueCount – the number of elements.
//
//	bitsPerValue – the number of bits available for any given value.
func NewPacked64(valueCount, bitsPerValue int) *Packed64 {
	format := FormatPacked
	longCount := format.LongCount(VERSION_CURRENT, valueCount, bitsPerValue)
	packed64 := &Packed64{blocks: make([]uint64, longCount)}

	packed64.maskRight = ^uint64(0) << (Packed64BlockSize - bitsPerValue) >> (Packed64BlockSize - bitsPerValue)
	packed64.bpvMinusBlockSize = bitsPerValue - Packed64BlockSize

	packed64.MutableImpl = newMutableImpl(packed64, valueCount, bitsPerValue)
	return packed64
}

func (p *Packed64) Get(index int) uint64 {
	// The abstract index in a bit stream
	majorBitPos := index * p.bitsPerValue

	// The index in the backing long-array
	elementPos := majorBitPos >> Packed64BlockBits

	// The number of value-bits in the second long
	endBits := (majorBitPos & Packed64ModMask) + p.bpvMinusBlockSize

	if endBits <= 0 {
		// Single block
		return (p.blocks[elementPos] >> -endBits) & p.maskRight
	}

	// Two blocks
	return ((p.blocks[elementPos] << endBits) |
		(p.blocks[elementPos+1] >> (Packed64BlockSize - endBits))) &
		p.maskRight
}

func (p *Packed64) GetBulk(index int, arr []uint64) int {
	off, length := 0, len(arr)
	length = min(length, p.valueCount-index)

	originalIndex := index

	decoder, err := Of(FormatPacked, p.bitsPerValue)
	if err != nil {
		return -1
	}

	// go to the next block where the value does not span across two blocks
	offsetInBlocks := index % decoder.LongBlockCount()
	if offsetInBlocks != 0 {
		for i := offsetInBlocks; i < decoder.LongValueCount() && length > 0; i++ {
			arr[off] = p.Get(index)
			off++
			index++
			length--
		}

		if length == 0 {
			return index - originalIndex
		}
	}

	// bulk get
	blockIndex := (index * p.bitsPerValue) >> Packed64BlockBits

	iterations := length / decoder.LongValueCount()
	decoder.DecodeLongToLong(p.blocks[blockIndex:], arr[off:], iterations)
	gotValues := iterations * decoder.LongValueCount()
	index += gotValues
	length -= gotValues

	if index > originalIndex {
		// stay at the block boundary
		return index - originalIndex
	}

	// no progress so far => already at a block boundary but no full block to get
	return GetBulk(p, index, arr[off:off+length])
}

func (p *Packed64) Set(index int, value uint64) {
	// The abstract index in a contiguous bit stream
	majorBitPos := index * p.bitsPerValue
	// The index in the backing long-array
	elementPos := majorBitPos >> Packed64BlockBits // / BLOCK_SIZE
	// The number of value-bits in the second long
	endBits := (majorBitPos & Packed64ModMask) + p.bpvMinusBlockSize

	if endBits <= 0 { // Single block
		p.blocks[elementPos] = p.blocks[elementPos] & ^(p.maskRight<<-endBits) | (value << -endBits)
		return
	}
	// Two blocks
	p.blocks[elementPos] = p.blocks[elementPos] & ^(p.maskRight>>endBits) | (value >> endBits)
	p.blocks[elementPos+1] = p.blocks[elementPos+1]&(^uint64(0)>>endBits) | (value << (Packed64BlockSize - endBits))
}

func (p *Packed64) SetBulk(index int, arr []uint64) int {
	off := 0
	size := min(len(arr), p.valueCount-index)
	originalIndex := index

	encoder, err := Of(FormatPacked, p.bitsPerValue)
	if err != nil {
		return -1
	}

	// go to the next block where the value does not span across two blocks
	offsetInBlocks := index % encoder.LongBlockCount()
	if offsetInBlocks != 0 {
		for i := offsetInBlocks; i < encoder.LongValueCount() && size > 0; i++ {
			p.Set(index, arr[off])

			index++
			off++
			size--
		}
		if size == 0 {
			return index - originalIndex
		}
	}

	blockIndex := (int)((index * p.bitsPerValue) >> Packed64BlockBits)

	iterations := size / encoder.LongValueCount()
	encoder.EncodeLongToLong(arr[off:], p.blocks[blockIndex:], iterations)
	setValues := iterations * encoder.LongValueCount()
	index += setValues
	size -= setValues

	if index > originalIndex {
		// stay at the block boundary
		return index - originalIndex
	}

	// no progress so far => already at a block boundary but no full block to get
	return SetBulk(p, index, arr[off:off+size])
}

func (p *Packed64) Fill(fromIndex, toIndex int, value uint64) {
	// minimum number of values that use an exact number of full blocks
	nAlignedValues := 64 / gcd(64, p.bitsPerValue)
	span := toIndex - fromIndex
	if span <= 3*nAlignedValues {
		// there needs be at least 2 * nAlignedValues aligned values for the
		// block approach to be worth trying
		Fill(p, fromIndex, toIndex, value)
		return
	}

	// fill the first values naively until the next block start
	fromIndexModNAlignedValues := fromIndex % nAlignedValues
	if fromIndexModNAlignedValues != 0 {
		for i := fromIndexModNAlignedValues; i < nAlignedValues; i++ {
			p.Set(fromIndex, value)
			fromIndex++
		}
	}

	// compute the long[] blocks for nAlignedValues consecutive values and
	// use them to set as many values as possible without applying any mask
	// or shift
	nAlignedBlocks := (nAlignedValues * p.bitsPerValue) >> 6
	nAlignedValuesBlocks := make([]uint64, 0)
	{
		values := NewPacked64(nAlignedValues, p.bitsPerValue)
		for i := 0; i < nAlignedValues; i++ {
			values.Set(i, value)
		}
		nAlignedValuesBlocks = values.blocks
	}
	startBlock := (fromIndex * p.bitsPerValue) >> 6
	endBlock := (toIndex * p.bitsPerValue) >> 6
	for block := startBlock; block < endBlock; block++ {
		blockValue := nAlignedValuesBlocks[block%nAlignedBlocks]
		p.blocks[block] = blockValue
	}

	// fill the gap
	for i := (endBlock << 6) / p.bitsPerValue; i < toIndex; i++ {
		p.Set(i, value)
	}
}

func gcd(a, b int) int {
	if a < b {
		return gcd(b, a)
	} else if b == 0 {
		return a
	} else {
		return gcd(b, a%b)
	}
}

func (p *Packed64) Clear() {
	for i := range p.blocks {
		p.blocks[i] = 0
	}
}

func (p *Packed64) Save(ctx context.Context, out store.DataOutput) error {
	writer := getWriterNoHeader(out, p.GetFormat(), p.Size(), p.GetBitsPerValue(), DEFAULT_BUFFER_SIZE)
	err := writer.WriteHeader(ctx)
	if err != nil {
		return err
	}
	for i := 0; i < p.Size(); i++ {
		err := writer.Add(p.Get(i))
		if err != nil {
			return err
		}
	}
	return writer.Finish()
}

func (p *Packed64) GetFormat() Format {
	return FormatPacked
}
