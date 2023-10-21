package ints

var _ IntsAllocator = &RecyclingIntBlockAllocator{}

type RecyclingIntBlockAllocator struct {
	*AllocatorImp

	freeByteBlocks          [][]int
	maxBufferedBlocks       int
	freeBlocks              int
	DEFAULT_BUFFERED_BLOCKS int
}

func NewRecyclingIntBlockAllocator(blockSize, maxBufferedBlocks int) *RecyclingIntBlockAllocator {
	allocator := RecyclingIntBlockAllocator{
		AllocatorImp:            nil,
		freeBlocks:              0,
		maxBufferedBlocks:       maxBufferedBlocks,
		DEFAULT_BUFFERED_BLOCKS: 64,
	}
	allocator.AllocatorImp = NewAllocator(blockSize, &allocator)
	return &allocator
}

func (r *RecyclingIntBlockAllocator) RecycleIntBlocks(blocks [][]int, start, end int) {
	panic("TODO")
}

func (r *RecyclingIntBlockAllocator) GetIntBlock() []int {
	if r.freeBlocks == 0 {
		return make([]int, r.blockSize)
	}
	b := r.freeByteBlocks[r.freeBlocks-1]
	r.freeBlocks--
	r.freeByteBlocks[r.freeBlocks] = nil
	return b
}
