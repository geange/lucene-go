package packed

import (
	"context"
	"errors"
	"fmt"
	"github.com/geange/lucene-go/core/store"
	"io"
)

const (
	Packed64BlockSize = 64                            // 32 = int, 64 = long
	Packed64BlockBits = 6                             // The #bits representing BLOCK_SIZE
	MOD_MASK          = uint64(Packed64BlockSize - 1) // x % BLOCK_SIZE
)

var _ Mutable = &Packed64{}

type Packed64 struct {
	*baseMutable

	blocks            []uint64 // Values are stores contiguously in the blocks array.
	maskRight         uint64   // A right-aligned mask of width BitsPerValue used by get(int).
	bpvMinusBlockSize int      // Optimization: Saves one lookup in get(int).
}

// NewPacked64
// Creates an array with the internal structures adjusted for the given limits and initialized to 0.
// valueCount: the number of elements.
// bitsPerValue: the number of bits available for any given value.
func NewPacked64(valueCount, bitsPerValue int) *Packed64 {
	format := FormatPacked
	longCount := format.LongCount(VERSION_CURRENT, valueCount, bitsPerValue)
	packed64 := &Packed64{blocks: make([]uint64, longCount)}

	packed64.maskRight = ^uint64(0) << (Packed64BlockSize - bitsPerValue) >> (Packed64BlockSize - bitsPerValue)
	packed64.bpvMinusBlockSize = bitsPerValue - Packed64BlockSize

	packed64.baseMutable = newBaseMutable(packed64, valueCount, bitsPerValue)
	return packed64
}

// NewPacked64V1
// Creates an array with content retrieved from the given DataInput.
// in: a DataInput, positioned at the start of Packed64-content.
// valueCount: the number of elements.
// bitsPerValue the number of bits available for any given value.
func NewPacked64V1(ctx context.Context, packedIntsVersion int, in store.DataInput, valueCount, bitsPerValue int) (*Packed64, error) {
	res := NewPacked64(valueCount, bitsPerValue)

	format := FormatPacked
	byteCount := format.ByteCount(packedIntsVersion, valueCount, bitsPerValue) // to know how much to read
	longCount := format.LongCount(VERSION_CURRENT, valueCount, bitsPerValue)   // to size the array
	res.blocks = make([]uint64, longCount)
	// read as many longs as we can
	for i := 0; i < byteCount/8; i++ {
		block, err := in.ReadUint64(ctx)
		if err != nil {
			return nil, err
		}
		res.blocks[i] = block
	}
	remaining := byteCount % 8
	if remaining != 0 {
		// read the last bytes
		lastLong := uint64(0)
		for i := 0; i < remaining; i++ {
			b, err := in.ReadByte()
			if err != nil {
				return nil, err
			}
			lastLong |= uint64(b) << (56 - i*8)
		}
		res.blocks[len(res.blocks)-1] = lastLong
	}

	res.maskRight = ^uint64(0) << (Packed64BlockSize - bitsPerValue) >> (Packed64BlockSize - bitsPerValue)
	res.bpvMinusBlockSize = bitsPerValue - Packed64BlockSize
	res.baseMutable = newBaseMutable(res, valueCount, bitsPerValue)
	return res, nil
}

func (p *Packed64) Get(index int) (uint64, error) {
	// The abstract index in a bit stream
	majorBitPos := uint64(index * p.bitsPerValue)

	// The index in the backing long-array
	elementPos := majorBitPos >> Packed64BlockBits

	if int(elementPos) >= len(p.blocks) {
		return 0, io.EOF
	}

	// The number of value-bits in the second long
	endBits := int(majorBitPos&MOD_MASK) + p.bpvMinusBlockSize

	if endBits <= 0 {
		// Single block
		return (p.blocks[elementPos] >> -endBits) & p.maskRight, nil
	}

	if int(elementPos+1) >= len(p.blocks) {
		return 0, io.EOF
	}

	// Two blocks
	return ((p.blocks[elementPos] << endBits) |
		(p.blocks[elementPos+1] >> (Packed64BlockSize - endBits))) &
		p.maskRight, nil
}

func (p *Packed64) GetTest(index int) uint64 {
	v, _ := p.Get(index)
	return v
}

func (p *Packed64) GetBulk(index int, buffer []uint64) int {
	bufferIndex, length := 0, len(buffer)
	length = min(length, p.valueCount-index)

	originalIndex := index

	decoder, err := Of(FormatPacked, p.bitsPerValue)
	if err != nil {
		return -1
	}

	// go to the next block where the value does not span across two blocks
	offsetInBlocks := index % decoder.LongValueCount()
	if offsetInBlocks != 0 {
		for i := offsetInBlocks; i < decoder.LongValueCount() && length > 0; i++ {
			n, err := p.Get(index)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return bufferIndex
				}
				return 0
			}
			buffer[bufferIndex] = n
			bufferIndex++
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
	decoder.DecodeUint64(p.blocks[blockIndex:], buffer[bufferIndex:], iterations)
	gotValues := iterations * decoder.LongValueCount()
	index += gotValues
	length -= gotValues

	if index > originalIndex {
		// stay at the block boundary
		return index - originalIndex
	}

	// no progress so far => already at a block boundary but no full block to get
	return GetBulk(p, index, buffer[bufferIndex:bufferIndex+length])
}

func (p *Packed64) Set(index int, value uint64) {
	// The abstract index in a contiguous bit stream
	majorBitPos := uint64(index * p.bitsPerValue)
	// The index in the backing long-array
	elementPos := majorBitPos >> Packed64BlockBits // / BLOCK_SIZE
	// The number of value-bits in the second long
	endBits := int(majorBitPos&MOD_MASK) + p.bpvMinusBlockSize

	if endBits <= 0 { // Single block
		if int(elementPos) >= len(p.blocks) {
			fmt.Println(elementPos, p.bitsPerValue)
		}
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

	blockIndex := (index * p.bitsPerValue) >> Packed64BlockBits

	iterations := size / encoder.LongValueCount()
	encoder.EncodeUint64(arr[off:], p.blocks[blockIndex:], iterations)
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
	var nAlignedValuesBlocks []uint64
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
		v, err := p.Get(i)
		if err != nil {
			return err
		}
		if err := writer.Add(v); err != nil {
			return err
		}
	}
	return writer.Finish()
}

func (p *Packed64) GetFormat() Format {
	return FormatPacked
}
