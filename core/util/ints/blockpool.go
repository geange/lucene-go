package ints

const (
	INT_BLOCK_SHIFT = 13
	INT_BLOCK_SIZE  = 1 << INT_BLOCK_SHIFT
	INT_BLOCK_MASK  = INT_BLOCK_SIZE - 1
)

// BlockPool A pool for int blocks similar to ByteBlockPool
type BlockPool struct {
	// array of buffers currently used in the pool. Buffers are allocated
	// if needed don't modify this outside of this class
	buffers [][]int

	// index into the buffers array pointing to the current buffer used as the head
	// 一级索引
	bufferUpto int

	// Pointer to the current position in head buffer
	// 二级索引
	intUpto int

	// Current head buffer
	// 当前的数组
	buffer []int

	// Current head offset
	intOffset int

	allocator IntsAllocator
}

//func NewBlockPool() *IntBlockPool {
//	return &IntBlockPool{
//		buffers:    make([][]int, 0, 10),
//		bufferUpto: -1,
//		intUpto:    INT_BLOCK_SIZE,
//		buffer:     make([]int, 0),
//		IntOffset:  -INT_BLOCK_SIZE,
//		allocator:  NewDirectAllocator(INT_BLOCK_SIZE),
//	}
//}

func NewBlockPool(allocator IntsAllocator) *BlockPool {
	pool := &BlockPool{
		buffers:    make([][]int, 0, 10),
		bufferUpto: -1,
		intUpto:    INT_BLOCK_SIZE,
		buffer:     make([]int, 0),
		intOffset:  -INT_BLOCK_SIZE,
		allocator:  allocator,
	}
	if allocator == nil {
		pool.allocator = NewDirectAllocator(INT_BLOCK_SIZE)
	}
	return pool
}

func (b *BlockPool) IntOffset() int {
	return b.intOffset
}

func (b *BlockPool) Get(index int) []int {
	return b.buffers[index]
}

func (b *BlockPool) IntUpto() int {
	return b.intUpto
}

func (b *BlockPool) AddIntUpto(v int) {
	b.intUpto += v
}

// Reset Expert: Resets the pool to its initial state reusing the first buffer.
// Params: 	zeroFillBuffers – if true the buffers are filled with 0. This should be set to true if this pool
//
//	is used with IntBlockPool.SliceWriter.
//	reuseFirst – if true the first buffer will be reused and calling nextBuffer() is not needed after
//	reset iff the block pool was used before ie. nextBuffer() was called before.
func (b *BlockPool) Reset(zeroFillBuffers, reuseFirst bool) {
	if b.bufferUpto != -1 {
		// We allocated at least one buffer
		if zeroFillBuffers {
			for idx := range b.buffers {
				// Fully zero fill buffers that we fully used
				for k := range b.buffers[idx] {
					b.buffers[idx][k] = 0
				}
			}
			// Partial zero fill the final buffer
		}

		if b.bufferUpto > 0 || !reuseFirst {
			offset := 0
			if reuseFirst {
				offset = 1
			}
			b.allocator.RecycleIntBlocks(b.buffers, offset, 1+b.bufferUpto)
			for idx := range b.buffers {
				if idx >= offset {
					b.buffers[idx] = nil
				}
			}
		}

		if reuseFirst {
			// Re-use the first buffer
			b.bufferUpto = 0
			b.intUpto = 0
			b.intOffset = 0
			b.buffer = b.buffers[0]
		} else {
			b.bufferUpto = -1
			b.intUpto = INT_BLOCK_SIZE
			b.intOffset = -INT_BLOCK_SIZE
			b.buffer = nil
		}
	}
}

// NextBuffer Advances the pool to its next buffer. This method should be called once after the constructor
// to initialize the pool. In contrast to the constructor a reset() call will advance the pool to its first
// buffer immediately.
func (b *BlockPool) NextBuffer() {
	b.buffers = append(b.buffers, b.allocator.GetIntBlock())
	b.buffer = b.buffers[b.bufferUpto+1]
	b.bufferUpto++
	b.intUpto = 0
	b.intOffset += INT_BLOCK_SIZE
}

// Creates a new int slice with the given starting size and returns the slices offset in the pool.
// See Also: IntBlockPool.SliceReader
func (b *BlockPool) newSlice(size int) int {
	if b.intUpto > INT_BLOCK_SIZE-size {
		b.NextBuffer()
	}

	upto := b.intUpto
	b.intUpto += size
	b.buffer[b.intUpto-1] = 16
	return upto
}

var (
	// INT_NEXT_LEVEL_ARRAY An array holding the offset into the INT_LEVEL_SIZE_ARRAY to quickly navigate to the next slice level.
	INT_NEXT_LEVEL_ARRAY = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 9}

	// INT_LEVEL_SIZE_ARRAY An array holding the level sizes for int slices.
	INT_LEVEL_SIZE_ARRAY = []int{2, 4, 8, 16, 16, 32, 32, 64, 64, 128}

	// INT_FIRST_LEVEL_SIZE The first level size for new slices
	INT_FIRST_LEVEL_SIZE = INT_LEVEL_SIZE_ARRAY[0]
)

// Allocates a new slice from the given offset
func (b *BlockPool) allocSlice(slice []int, sliceOffset int) int {
	level := slice[sliceOffset] & 15
	newLevel := INT_NEXT_LEVEL_ARRAY[level]
	newSize := INT_LEVEL_SIZE_ARRAY[newLevel]
	if b.intUpto > INT_BLOCK_SIZE-newSize {
		b.NextBuffer()
	}

	newUpto := b.intUpto
	offset := newUpto + b.intOffset
	b.intUpto += newSize
	// Write forwarding address at end of last slice:
	slice[sliceOffset] = offset

	// Write new level:
	b.buffer[b.intUpto-1] = 16 | newLevel

	return newUpto
}

func (b *BlockPool) Buffer() []int {
	return b.buffer
}

// SliceWriter A IntBlockPool.SliceWriter that allows to write multiple integer slices into a given BlockPool.
// See Also: IntBlockPool.SliceReader
type SliceWriter struct {
	offset int
	pool   *BlockPool
}

func NewSliceWriter(pool *BlockPool) *SliceWriter {
	return &SliceWriter{pool: pool}
}

func (s *SliceWriter) Reset(sliceOffset int) {
	s.offset = sliceOffset
}

func (s *SliceWriter) WriteInt(value int) {
	ints := s.pool.buffers[s.offset>>INT_BLOCK_SHIFT]
	relativeOffset := s.offset & INT_BLOCK_MASK
	if ints[relativeOffset] != 0 {
		// End of slice; allocate a new one

		relativeOffset = s.pool.allocSlice(ints, relativeOffset)
		ints = s.pool.buffer
		s.offset = relativeOffset + s.pool.intOffset
	}

	ints[relativeOffset] = value
	s.offset++
}

// StartNewSlice starts a new slice and returns the start offset. The returned value should be used as the
// start offset to initialize a IntBlockPool.SliceReader.
func (s *SliceWriter) StartNewSlice() int {
	s.offset = s.pool.newSlice(INT_FIRST_LEVEL_SIZE) + s.pool.intOffset
	return s.offset
}

// GetCurrentOffset Returns the offset of the currently written slice. The returned value should be used as the
// end offset to initialize a IntBlockPool.SliceReader once this slice is fully written or to reset the this
// writer if another slice needs to be written.
func (s *SliceWriter) GetCurrentOffset() int {
	return s.offset
}

type IntsAllocator interface {
	RecycleIntBlocks(blocks [][]int, start, end int)
	GetIntBlock() []int
}

type IntsArrayAllocatorNeed interface {
	RecycleIntBlocks(blocks [][]int, start, end int)
}

var _ IntsAllocator = &IntsAllocatorDefault{}

type IntsAllocatorDefault struct {
	BlockSize          int
	FnRecycleIntBlocks func(blocks [][]int, start, end int)
}

func (i *IntsAllocatorDefault) RecycleIntBlocks(blocks [][]int, start, end int) {
	i.FnRecycleIntBlocks(blocks, start, end)
}

func (i *IntsAllocatorDefault) GetIntBlock() []int {
	return make([]int, i.BlockSize)
}

// AllocatorImp Abstract class for allocating and freeing int blocks.
type AllocatorImp struct {
	blockSize int

	IntsArrayAllocatorNeed
}

func NewAllocator(blockSize int, need IntsArrayAllocatorNeed) *AllocatorImp {
	return &AllocatorImp{
		blockSize:              blockSize,
		IntsArrayAllocatorNeed: need,
	}
}

//func (a *IntsAllocatorImp) RecycleIntBlocks(blocks [][]int, start, end int) {
//	a.RecycleIntBlocks(blocks, start, end)
//}

func (a *AllocatorImp) GetIntBlock() []int {
	return make([]int, a.blockSize)
}

type DirectAllocator struct {
	*AllocatorImp
}

func NewDirectAllocator(blockSize int) *DirectAllocator {
	allocator := &DirectAllocator{}
	allocator.AllocatorImp = NewAllocator(blockSize, allocator)
	return allocator
}

func (d *DirectAllocator) RecycleIntBlocks(blocks [][]int, start, end int) {
}

// SliceReader A IntBlockPool.SliceReader that can read int slices written by a IntBlockPool.SliceWriter
type SliceReader struct {
	pool         *BlockPool
	upto         int
	bufferUpto   int
	bufferOffset int
	buffer       []int
	limit        int
	level        int
	end          int
}

func NewSliceReader(pool *BlockPool) *SliceReader {
	return &SliceReader{pool: pool}
}

// Reset Resets the reader to a slice give the slices absolute start and end offset in the pool
func (s *SliceReader) Reset(startOffset, endOffset int) {
	s.bufferUpto = startOffset / INT_BLOCK_SIZE
	s.bufferOffset = s.bufferUpto * INT_BLOCK_SIZE
	s.end = endOffset
	s.level = 0

	s.buffer = s.pool.buffers[s.bufferUpto]
	s.upto = startOffset & INT_BLOCK_MASK

	firstSize := INT_LEVEL_SIZE_ARRAY[0]
	if startOffset+firstSize >= endOffset {
		// There is only this one slice to read
		s.limit = endOffset & INT_BLOCK_MASK
	} else {
		s.limit = s.upto + firstSize - 1
	}
}

// EndOfSlice Returns true iff the current slice is fully read. If this method returns true readInt()
// should not be called again on this slice.
func (s *SliceReader) EndOfSlice() bool {
	return s.upto+s.bufferOffset == s.end
}

// ReadInt Reads the next int from the current slice and returns it.
// See Also: endOfSlice()
func (s *SliceReader) ReadInt() int {
	if s.upto == s.limit {
		s.nextSlice()
	}

	upto := s.upto
	s.upto++
	return s.buffer[upto]
}

func (s *SliceReader) nextSlice() {
	// Skip to our next slice
	nextIndex := s.buffer[s.limit]
	s.level = INT_NEXT_LEVEL_ARRAY[s.level]
	newSize := INT_LEVEL_SIZE_ARRAY[s.level]

	s.bufferUpto = nextIndex / INT_BLOCK_SIZE
	s.bufferOffset = s.bufferUpto * INT_BLOCK_SIZE

	s.buffer = s.pool.buffers[s.bufferUpto]
	s.upto = nextIndex & INT_BLOCK_MASK

	if nextIndex+newSize >= s.end {
		// We are advancing to the final slice
		s.limit = s.end - s.bufferOffset
	} else {
		// This is not the final slice (subtract 4 for the
		// forwarding address at the end of this new slice)
		s.limit = s.upto + newSize - 1
	}
}
