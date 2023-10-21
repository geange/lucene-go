package bytesutils

import (
	"slices"
	"sync/atomic"
)

// Allocator Abstract class for allocating and freeing byte blocks.
type Allocator interface {
	RecycleByteBlocks(blocks [][]byte, start, end int)
	GetByteBlock() []byte
}

var _ Allocator = &BytesAllocator{}

type BytesAllocator struct {
	BlockSize                int
	FnRecycleByteBlocksRange func(blocks [][]byte, start, end int)
}

func (b *BytesAllocator) RecycleByteBlocks(blocks [][]byte, start, end int) {
	b.FnRecycleByteBlocksRange(blocks, start, end)
}

func (b *BytesAllocator) GetByteBlock() []byte {
	return make([]byte, b.BlockSize)
}

var _ Allocator = &DirectAllocator{}

type DirectAllocator struct {
	blockSize int
}

func NewDirectAllocator(blockSize int) *DirectAllocator {
	return &DirectAllocator{blockSize: blockSize}
}

func (d *DirectAllocator) RecycleByteBlocks(blocks [][]byte, start, end int) {}

func (d *DirectAllocator) GetByteBlock() []byte {
	return make([]byte, d.blockSize)
}

type DirectTrackingAllocator struct {
}

const (
	DEFAULT_BUFFERED_BLOCKS = 64
)

var _ Allocator = &RecyclingByteBlockAllocator{}

type RecyclingByteBlockAllocator struct {
	blockSize         int
	freeByteBlocks    [][]byte
	maxBufferedBlocks int
	freeBlocks        int
	bytesUsed         *atomic.Int64
}

func NewRecyclingByteBlockAllocator(blockSize, maxBufferedBlocks int) *RecyclingByteBlockAllocator {
	allocator := RecyclingByteBlockAllocator{
		blockSize:         blockSize,
		freeBlocks:        0,
		maxBufferedBlocks: maxBufferedBlocks,
		bytesUsed:         new(atomic.Int64),
	}
	return &allocator
}

func (r *RecyclingByteBlockAllocator) GetByteBlock() []byte {
	if r.freeBlocks == 0 {
		r.bytesUsed.Add(int64(r.blockSize))
		return make([]byte, r.blockSize)
	}
	b := r.freeByteBlocks[r.freeBlocks-1]
	r.freeBlocks--
	r.freeByteBlocks[r.freeBlocks] = nil
	return b
}

func (r *RecyclingByteBlockAllocator) RecycleByteBlocks(blocks [][]byte, start, end int) {
	numBlocks := min(r.maxBufferedBlocks-r.freeBlocks, end-start)
	size := r.freeBlocks + numBlocks
	if size >= len(r.freeByteBlocks) {
		r.freeByteBlocks = slices.Clone(r.freeByteBlocks[:r.freeBlocks])
	}
	stop := start + numBlocks
	for i := start; i < stop; i++ {
		r.freeByteBlocks[r.freeBlocks] = blocks[i]
		r.freeBlocks++
		blocks[i] = nil
	}

	for i := stop; i < end; i++ {
		blocks[i] = nil
	}
	r.bytesUsed.Add(int64(-(end - stop) * r.blockSize))
	return
}
