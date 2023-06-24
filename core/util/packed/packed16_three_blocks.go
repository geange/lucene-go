package packed

import (
	"math"
)

const (
	Packed16ThreeBlocksMaxSize = math.MaxInt32 / 3
)

var _ Mutable = &Packed16ThreeBlocks{}

// Packed16ThreeBlocks Packs integers into 3 shorts (48 bits per value).
// lucene.internal
type Packed16ThreeBlocks struct {
	*MutableImpl

	blocks []uint16
}

func NewPacked16ThreeBlocks(valueCount int) *Packed16ThreeBlocks {
	blocks := &Packed16ThreeBlocks{blocks: make([]uint16, valueCount*3)}
	blocks.MutableImpl = newMutableImpl(blocks, valueCount, 48)
	return blocks
}

func (p *Packed16ThreeBlocks) Get(index int) uint64 {
	o := index * 3
	return uint64(p.blocks[o]&0xFFFF)<<32 |
		uint64(p.blocks[o+1]&0xFFFF)<<16 |
		uint64(p.blocks[o+2]&0xFFFF)
}

func (p *Packed16ThreeBlocks) GetBulk(index int, arr []uint64) int {
	gets := min(p.valueCount-index, len(arr))
	end := (index + gets) * 3

	off := 0
	for i := index * 3; i < end; i += 3 {
		arr[off] = uint64(p.blocks[i]&0xFFFF)<<32 |
			uint64(p.blocks[i+1]&0xFFFF)<<16 |
			uint64(p.blocks[i+2]&0xFFFF)
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

	for i := 0; i < sets; i++ {
		off := index * 3
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
