package packed

import (
	. "github.com/geange/lucene-go/math"
	"math"
)

const (
	Packed8ThreeBlocksMaxSize = math.MaxInt32 / 3
)

var _ Mutable = &Packed8ThreeBlocks{}

// Packed8ThreeBlocks Packs integers into 3 bytes (24 bits per value).
// lucene.internal
type Packed8ThreeBlocks struct {
	*MutableImpl

	blocks []byte
}

func NewPacked8ThreeBlocks(valueCount int) *Packed8ThreeBlocks {
	blocks := &Packed8ThreeBlocks{blocks: make([]byte, valueCount*3)}
	blocks.MutableImpl = newMutableImpl(blocks, valueCount, 24)
	return blocks
}

func (p *Packed8ThreeBlocks) Get(index int) uint64 {
	o := index * 3
	return uint64(p.blocks[o]&0xFF)<<16 |
		uint64(p.blocks[o+1]&0xFF)<<8 |
		uint64(p.blocks[o+2])
}

func (p *Packed8ThreeBlocks) GetBulk(index int, arr []uint64) int {
	gets := Min(p.valueCount-index, len(arr))
	end := (index + gets) * 3

	off := 0
	for i := index * 3; i < end; i += 3 {
		arr[off] = uint64(p.blocks[i]&0xFF)<<16 |
			uint64(p.blocks[i+1]&0xFF)<<16 |
			uint64(p.blocks[i+2]&0xFF)
		off++
	}
	return gets
}

func (p *Packed8ThreeBlocks) Set(index int, value uint64) {
	off := index * 3
	p.blocks[off] = byte(value >> 16)
	p.blocks[off+1] = byte(value >> 8)
	p.blocks[off+2] = byte(value)
}

func (p *Packed8ThreeBlocks) SetBulk(index int, arr []uint64) int {
	sets := Min(p.valueCount-index, len(arr))

	for i, off := 0, index*3; i < sets; i++ {
		value := arr[i]
		p.blocks[off] = byte(value >> 16)
		p.blocks[off+1] = byte(value >> 8)
		p.blocks[off+2] = byte(value)
		off += 3
	}
	return sets
}

func (p *Packed8ThreeBlocks) Fill(fromIndex, toIndex int, value uint64) {
	block1 := byte(value >> 16)
	block2 := byte(value >> 8)
	block3 := byte(value)

	end := toIndex * 3

	for i := fromIndex * 3; i < end; i += 3 {
		p.blocks[i] = block1
		p.blocks[i+1] = block2
		p.blocks[i+2] = block3
	}
}

func (p *Packed8ThreeBlocks) Clear() {
	for i := range p.blocks {
		p.blocks[i] = 0
	}
}