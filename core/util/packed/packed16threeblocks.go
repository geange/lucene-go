package packed

import (
	"context"
	"math"

	"github.com/geange/lucene-go/core/store"
)

const (
	Packed16ThreeBlocks_MAX_SIZE = math.MaxInt32 / 3
)

var _ Mutable = &Packed16ThreeBlocks{}

// Packed16ThreeBlocks Packs integers into 3 shorts (48 bits per value).
// lucene.internal
type Packed16ThreeBlocks struct {
	*baseMutable

	blocks []uint16
}

func NewPacked16ThreeBlocks(valueCount int) *Packed16ThreeBlocks {
	blocks := &Packed16ThreeBlocks{blocks: make([]uint16, valueCount*3)}
	blocks.baseMutable = newBaseMutable(blocks, valueCount, 48)
	return blocks
}

func NewPacked16ThreeBlocksV1(ctx context.Context, packedIntsVersion int,
	in store.DataInput, valueCount int) (*Packed16ThreeBlocks, error) {

	blocks := NewPacked16ThreeBlocks(valueCount)
	for i := 0; i < 3*valueCount; i++ {
		block, err := in.ReadUint16(ctx)
		if err != nil {
			return nil, err
		}
		blocks.blocks[i] = block
	}
	return blocks, nil
}

func (p *Packed16ThreeBlocks) Get(index int) (uint64, error) {
	o := index * 3
	return uint64(p.blocks[o])<<32 |
		uint64(p.blocks[o+1])<<16 |
		uint64(p.blocks[o+2]), nil
}

func (p *Packed16ThreeBlocks) GetTest(index int) uint64 {
	v, _ := p.Get(index)
	return v
}

func (p *Packed16ThreeBlocks) GetBulk(index int, arr []uint64) int {
	gets := min(p.valueCount-index, len(arr))
	end := (index + gets) * 3

	off := 0
	for i := index * 3; i < end; i += 3 {
		arr[off] = uint64(p.blocks[i])<<32 |
			uint64(p.blocks[i+1])<<16 |
			uint64(p.blocks[i+2])
		off++
	}
	return gets
}

func (p *Packed16ThreeBlocks) Set(index int, value uint64) {
	off := index * 3
	p.blocks[off] = uint16(value >> 32)
	p.blocks[off+1] = uint16(value >> 16)
	p.blocks[off+2] = uint16(value)
}

func (p *Packed16ThreeBlocks) SetBulk(index int, arr []uint64) int {
	sets := min(p.valueCount-index, len(arr))

	for i, off := 0, index*3; i < sets; i++ {
		value := arr[i]
		p.blocks[off] = uint16(value >> 32)
		p.blocks[off+1] = uint16(value >> 16)
		p.blocks[off+2] = uint16(value)
		off += 3
	}

	return sets
}

func (p *Packed16ThreeBlocks) Fill(fromIndex, toIndex int, value uint64) {
	block1 := uint16(value >> 32)
	block2 := uint16(value >> 16)
	block3 := uint16(value)

	for i := fromIndex * 3; i < toIndex*3; i += 3 {
		p.blocks[i] = block1
		p.blocks[i+1] = block2
		p.blocks[i+2] = block3
	}
}

func (p *Packed16ThreeBlocks) Clear() {
	for i := range p.blocks {
		p.blocks[i] = 0
	}
}
