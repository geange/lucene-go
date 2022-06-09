package util

import "go.uber.org/atomic"

type RecyclingByteBlockAllocator struct {
	*BytesAllocator

	freeByteBlocks          [][]byte
	maxBufferedBlocks       int
	freeBlocks              int
	bytesUsed               *atomic.Int64
	DEFAULT_BUFFERED_BLOCKS int
}

func NewRecyclingByteBlockAllocator(blockSize, maxBufferedBlocks int,
	bytesUsed *atomic.Int64) *RecyclingByteBlockAllocator {

	allocator := NewBytesAllocator(blockSize, nil)
	res := NewRecyclingByteBlockAllocatorDefault(allocator)
	res.maxBufferedBlocks = maxBufferedBlocks
	res.bytesUsed = bytesUsed
	return res
}

func NewRecyclingByteBlockAllocatorDefault(bytesAllocator *BytesAllocator) *RecyclingByteBlockAllocator {
	allocator := RecyclingByteBlockAllocator{
		BytesAllocator:          bytesAllocator,
		freeBlocks:              0,
		bytesUsed:               atomic.NewInt64(0),
		DEFAULT_BUFFERED_BLOCKS: 64,
	}
	bytesAllocator.ext = &allocator
	return &allocator
}

func (r *RecyclingByteBlockAllocator) RecycleByteBlocks(blocks [][]byte, start, end int) {
	//numBlocks := Min(r.maxBufferedBlocks-r.freeBlocks, end-start)
	//size := r.freeBlocks + numBlocks
	//if size >= len(r.freeByteBlocks) {
	//	newBlocks := make([][]byte)
	//}
	return
}
