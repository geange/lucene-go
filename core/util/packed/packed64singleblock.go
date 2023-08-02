package packed

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/store"
)

type Packed64SingleBlock struct {
	*baseMutable

	blocks []uint64
}

func NewPacked64SingleBlock(valueCount, bitsPerValue int) (*Packed64SingleBlock, error) {
	switch bitsPerValue {
	case 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 16, 21, 32:
		valuesPerBlock := 64 / bitsPerValue
		block := &Packed64SingleBlock{blocks: make([]uint64, requiredCapacity(valueCount, valuesPerBlock))}
		block.baseMutable = newBaseMutable(block, valueCount, bitsPerValue)
		return block, nil
	default:
		return nil, errors.New("unsupported number of bits per value")
	}
}

func NewPacked64SingleBlockV1(ctx context.Context, in store.DataInput, valueCount, bitsPerValue int) (*Packed64SingleBlock, error) {
	reader, err := NewPacked64SingleBlock(valueCount, bitsPerValue)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(reader.blocks); i++ {
		block, err := in.ReadUint64(ctx)
		if err != nil {
			return nil, err
		}
		reader.blocks[i] = block
	}
	return reader, nil
}

func (p *Packed64SingleBlock) Get(index int) (uint64, error) {
	switch p.bitsPerValue {
	case 1:
		o := index >> 6
		b := index & 63
		shift := b << 0
		return (p.blocks[o] >> shift) & 1, nil
	case 2:
		o := index >> 5
		b := index & 31
		shift := b << 1
		return (p.blocks[o] >> shift) & 3, nil
	case 3:
		o := index / 21
		b := index % 21
		shift := b * 3
		return (p.blocks[o] >> shift) & 7, nil
	case 4:
		o := index >> 4
		b := index & 15
		shift := b << 2
		return (p.blocks[o] >> shift) & 15, nil
	case 5:
		o := index / 12
		b := index % 12
		shift := b * 5
		return (p.blocks[o] >> shift) & 31, nil
	case 6:
		o := index / 10
		b := index % 10
		shift := b * 6
		return (p.blocks[o] >> shift) & 63, nil
	case 7:
		o := index / 9
		b := index % 9
		shift := b * 7
		return (p.blocks[o] >> shift) & 127, nil
	case 8:
		o := index >> 3
		b := index & 7
		shift := b << 3
		return (p.blocks[o] >> shift) & 255, nil
	case 9:
		o := index / 7
		b := index % 7
		shift := b * 9
		return (p.blocks[o] >> shift) & 511, nil
	case 10:
		o := index / 6
		b := index % 6
		shift := b * 10
		return (p.blocks[o] >> shift) & 1023, nil
	case 12:
		o := index / 5
		b := index % 5
		shift := b * 12
		return (p.blocks[o] >> shift) & 4095, nil
	case 16:
		o := index >> 2
		b := index & 3
		shift := b << 4
		return (p.blocks[o] >> shift) & 65535, nil
	case 21:
		o := index / 3
		b := index % 3
		shift := b * 21
		return (p.blocks[o] >> shift) & 2097151, nil
	case 32:
		o := index >> 1
		b := index & 1
		shift := b << 5
		return (p.blocks[o] >> shift) & 4294967295, nil
	}

	return 0, nil
}

func (p *Packed64SingleBlock) Set(index int, value uint64) {
	switch p.bitsPerValue {
	case 1:
		o := index >> 6
		b := index & 63
		shift := b << 0
		p.blocks[o] = (p.blocks[o] & ^(1 << shift)) | (value << shift)
	case 2:
		o := index >> 5
		b := index & 31
		shift := b << 1
		p.blocks[o] = (p.blocks[o] & ^(3 << shift)) | (value << shift)
	case 3:
		o := index / 21
		b := index % 21
		shift := b * 3
		p.blocks[o] = (p.blocks[o] & ^(7 << shift)) | (value << shift)
	case 4:
		o := index >> 4
		b := index & 15
		shift := b << 2
		p.blocks[o] = (p.blocks[o] & ^(15 << shift)) | (value << shift)
	case 5:
		o := index / 12
		b := index % 12
		shift := b * 5
		p.blocks[o] = (p.blocks[o] & ^(31 << shift)) | (value << shift)
	case 6:
		o := index / 10
		b := index % 10
		shift := b * 6
		p.blocks[o] = (p.blocks[o] & ^(63 << shift)) | (value << shift)
	case 7:
		o := index / 9
		b := index % 9
		shift := b * 7
		p.blocks[o] = (p.blocks[o] & ^(127 << shift)) | (value << shift)
	case 8:
		o := index >> 3
		b := index & 7
		shift := b << 3
		p.blocks[o] = (p.blocks[o] & ^(255 << shift)) | (value << shift)
	case 9:
		o := index / 7
		b := index % 7
		shift := b * 9
		p.blocks[o] = (p.blocks[o] & ^(511 << shift)) | (value << shift)
	case 10:
		o := index / 6
		b := index % 6
		shift := b * 10
		p.blocks[o] = (p.blocks[o] & ^(1023 << shift)) | (value << shift)
	case 12:
		o := index / 5
		b := index % 5
		shift := b * 12
		p.blocks[o] = (p.blocks[o] & ^(4095 << shift)) | (value << shift)
	case 16:
		o := index >> 2
		b := index & 3
		shift := b << 4
		p.blocks[o] = (p.blocks[o] & ^(65535 << shift)) | (value << shift)
	case 21:
		o := index / 3
		b := index % 3
		shift := b * 21
		p.blocks[o] = (p.blocks[o] & ^(2097151 << shift)) | (value << shift)
	case 32:
		o := index >> 1
		b := index & 1
		shift := b << 5
		p.blocks[o] = (p.blocks[o] & ^(4294967295 << shift)) | (value << shift)
	}

}

func (p *Packed64SingleBlock) Clear() {
	clear(p.blocks)
}

var (
	//SUPPORTED_BITS_PER_VALUE = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 16, 21, 32}
	SUPPORTED_BITS_PER_VALUE = map[int]bool{
		1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true, 9: true, 10: true, 12: true, 16: true, 21: true, 32: true,
	}
)

func isSupported(bitsPerValue int) bool {
	return SUPPORTED_BITS_PER_VALUE[bitsPerValue]
}

func requiredCapacity(valueCount, valuesPerBlock int) int {
	add := 1
	if valueCount%valuesPerBlock == 0 {
		add = 0
	}
	return valueCount/valuesPerBlock + add
}
