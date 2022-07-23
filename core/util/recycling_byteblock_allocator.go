package util

type RecyclingByteBlockAllocator struct {
	*BytesAllocatorImp

	freeByteBlocks          [][]byte
	maxBufferedBlocks       int
	freeBlocks              int
	DEFAULT_BUFFERED_BLOCKS int
}

func NewRecyclingByteBlockAllocator(blockSize, maxBufferedBlocks int) *RecyclingByteBlockAllocator {
	allocator := RecyclingByteBlockAllocator{
		BytesAllocatorImp:       nil,
		freeBlocks:              0,
		maxBufferedBlocks:       maxBufferedBlocks,
		DEFAULT_BUFFERED_BLOCKS: 64,
	}
	allocator.BytesAllocatorImp = NewBytesAllocator(blockSize, &allocator)
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
