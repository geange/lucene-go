package bytesref

import (
	"slices"
	"sync/atomic"
)

var gAllocatorBuilder = &AllocatorBuilder{}

func GetAllocatorBuilder() *AllocatorBuilder {
	return gAllocatorBuilder
}

type AllocatorBuilder struct {
}

func (b *AllocatorBuilder) NewDirect(blockSize int) Allocator {
	return newDirectAllocator(blockSize)
}

func (b *AllocatorBuilder) NewRecyclingByteBlock(blockSize, maxBufferedBlocks int) Allocator {
	return newRecyclingByteBlockAllocator(blockSize, maxBufferedBlocks)
}

func (b *AllocatorBuilder) NewBytes(blockSize int, fn func(blocks [][]byte, start, end int)) Allocator {
	return &bytesAllocator{
		blockSize:           blockSize,
		fnRecycleByteBlocks: fn,
	}
}

// Allocator Abstract class for allocating and freeing byte blocks.
type Allocator interface {
	RecycleByteBlocks(blocks [][]byte, start, end int)
	GetByteBlock() []byte
}

var _ Allocator = &bytesAllocator{}

type bytesAllocator struct {
	blockSize           int
	fnRecycleByteBlocks func(blocks [][]byte, start, end int)
}

func (b *bytesAllocator) RecycleByteBlocks(blocks [][]byte, start, end int) {
	b.fnRecycleByteBlocks(blocks, start, end)
}

func (b *bytesAllocator) GetByteBlock() []byte {
	return make([]byte, b.blockSize)
}

var _ Allocator = &directAllocator{}

type directAllocator struct {
	blockSize int
}

func newDirectAllocator(blockSize int) *directAllocator {
	return &directAllocator{blockSize: blockSize}
}

func (d *directAllocator) RecycleByteBlocks(blocks [][]byte, start, end int) {}

func (d *directAllocator) GetByteBlock() []byte {
	return make([]byte, d.blockSize)
}

type DirectTrackingAllocator struct {
}

const (
	DEFAULT_BUFFERED_BLOCKS = 64
)

var _ Allocator = &recyclingByteBlockAllocator{}

type recyclingByteBlockAllocator struct {
	blockSize         int
	freeByteBlocks    [][]byte
	maxBufferedBlocks int
	freeBlocks        int
	bytesUsed         *atomic.Int64
}

func newRecyclingByteBlockAllocator(blockSize, maxBufferedBlocks int) *recyclingByteBlockAllocator {
	allocator := recyclingByteBlockAllocator{
		blockSize:         blockSize,
		freeBlocks:        0,
		maxBufferedBlocks: maxBufferedBlocks,
		bytesUsed:         new(atomic.Int64),
	}
	return &allocator
}

func (r *recyclingByteBlockAllocator) GetByteBlock() []byte {
	if r.freeBlocks == 0 {
		r.bytesUsed.Add(int64(r.blockSize))
		return make([]byte, r.blockSize)
	}
	b := r.freeByteBlocks[r.freeBlocks-1]
	r.freeBlocks--
	r.freeByteBlocks[r.freeBlocks] = nil
	return b
}

func (r *recyclingByteBlockAllocator) RecycleByteBlocks(blocks [][]byte, start, end int) {
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
