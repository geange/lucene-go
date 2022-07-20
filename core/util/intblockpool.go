package util

const (
	INT_BLOCK_SHIFT = 13
	INT_BLOCK_SIZE  = 1 << INT_BLOCK_SHIFT
	INT_BLOCK_MASK  = INT_BLOCK_SIZE - 1
)

// IntBlockPool A pool for int blocks similar to ByteBlockPool
type IntBlockPool struct {
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

	allocator *IntsAllocator
}

func NewIntBlockPool() *IntBlockPool {
	return &IntBlockPool{
		buffers:    make([][]int, 0, 10),
		bufferUpto: -1,
		intUpto:    INT_BLOCK_SIZE,
		buffer:     make([]int, 0),
		intOffset:  -INT_BLOCK_SIZE,
		allocator:  NewIntsAllocator(INT_BLOCK_SIZE, &DirectIntsAllocator{}),
	}
}

// Reset Expert: Resets the pool to its initial state reusing the first buffer.
// Params: 	zeroFillBuffers – if true the buffers are filled with 0. This should be set to true if this pool
//			is used with IntBlockPool.SliceWriter.
//			reuseFirst – if true the first buffer will be reused and calling nextBuffer() is not needed after
//			reset iff the block pool was used before ie. nextBuffer() was called before.
func (i *IntBlockPool) Reset(zeroFillBuffers, reuseFirst bool) {
	if i.bufferUpto != -1 {
		// We allocated at least one buffer
		if zeroFillBuffers {
			for idx := range i.buffers {
				// Fully zero fill buffers that we fully used
				for k := range i.buffers[idx] {
					i.buffers[idx][k] = 0
				}
			}
			// Partial zero fill the final buffer
		}

		if i.bufferUpto > 0 || !reuseFirst {
			offset := 0
			if reuseFirst {
				offset = 1
			}
			i.allocator.RecycleIntBlocks(i.buffers, offset, 1+i.bufferUpto)
			for idx := range i.buffers {
				if idx >= offset {
					i.buffers[idx] = nil
				}
			}
		}

		if reuseFirst {
			// Re-use the first buffer
			i.bufferUpto = 0
			i.intUpto = 0
			i.intOffset = 0
			i.buffer = i.buffers[0]
		} else {
			i.bufferUpto = -1
			i.intUpto = INT_BLOCK_SIZE
			i.intOffset = -INT_BLOCK_SIZE
			i.buffer = nil
		}
	}
}

// NextBuffer Advances the pool to its next buffer. This method should be called once after the constructor
// to initialize the pool. In contrast to the constructor a reset() call will advance the pool to its first
// buffer immediately.
func (i *IntBlockPool) NextBuffer() {
	i.buffers = append(i.buffers, i.allocator.GetIntBlock())
	i.buffer = i.buffers[i.bufferUpto+1]
	i.bufferUpto++
	i.intUpto = 0
	i.intOffset += INT_BLOCK_SIZE
}

// Creates a new int slice with the given starting size and returns the slices offset in the pool.
// See Also: IntBlockPool.SliceReader
func (i *IntBlockPool) newSlice(size int) int {
	if i.intUpto > INT_BLOCK_SIZE-size {
		i.NextBuffer()
	}

	upto := i.intUpto
	i.intUpto += size
	i.buffer[i.intUpto-1] = 16
	return upto
}

var (
	// NEXT_LEVEL_ARRAY An array holding the offset into the LEVEL_SIZE_ARRAY to quickly navigate to the next slice level.
	NEXT_LEVEL_ARRAY = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 9}

	// LEVEL_SIZE_ARRAY An array holding the level sizes for int slices.
	LEVEL_SIZE_ARRAY = []int{2, 4, 8, 16, 16, 32, 32, 64, 64, 128}

	// FIRST_LEVEL_SIZE The first level size for new slices
	FIRST_LEVEL_SIZE = LEVEL_SIZE_ARRAY[0]
)

// Allocates a new slice from the given offset
func (i *IntBlockPool) allocSlice(slice []int, sliceOffset int) int {
	level := slice[sliceOffset] & 15
	newLevel := NEXT_LEVEL_ARRAY[level]
	newSize := LEVEL_SIZE_ARRAY[newLevel]
	if i.intUpto > INT_BLOCK_SIZE-newSize {
		i.NextBuffer()
	}

	newUpto := i.intUpto
	offset := newUpto + i.intOffset
	i.intUpto += newSize
	// Write forwarding address at end of last slice:
	slice[sliceOffset] = offset

	// Write new level:
	i.buffer[i.intUpto-1] = 16 | newLevel

	return newUpto
}

// SliceWriter A IntBlockPool.SliceWriter that allows to write multiple integer slices into a given IntBlockPool.
// See Also: IntBlockPool.SliceReader
type SliceWriter struct {
	offset int
	pool   *IntBlockPool
}

func NewSliceWriter(pool *IntBlockPool) *SliceWriter {
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
	s.offset = s.pool.newSlice(FIRST_LEVEL_SIZE) + s.pool.intOffset
	return s.offset
}

// GetCurrentOffset Returns the offset of the currently written slice. The returned value should be used as the
// end offset to initialize a IntBlockPool.SliceReader once this slice is fully written or to reset the this
// writer if another slice needs to be written.
func (s *SliceWriter) GetCurrentOffset() int {
	return s.offset
}

// IntsAllocator Abstract class for allocating and freeing int blocks.
type IntsAllocator struct {
	blockSize int
	ext       IntArrayAllocatorExt
}

func NewIntsAllocator(blockSize int, ext IntArrayAllocatorExt) *IntsAllocator {
	return &IntsAllocator{
		blockSize: blockSize,
		ext:       ext,
	}
}

func (a *IntsAllocator) RecycleIntBlocks(blocks [][]int, start, end int) {
	a.ext.RecycleIntBlocks(blocks, start, end)
}

func (a *IntsAllocator) GetIntBlock() []int {
	return make([]int, a.blockSize)
}

type IntArrayAllocatorExt interface {
	RecycleIntBlocks(blocks [][]int, start, end int)
}

type DirectIntsAllocator struct {
}

func (d *DirectIntsAllocator) RecycleIntBlocks(blocks [][]int, start, end int) {
}

// SliceReader A IntBlockPool.SliceReader that can read int slices written by a IntBlockPool.SliceWriter
type SliceReader struct {
	pool         *IntBlockPool
	upto         int
	bufferUpto   int
	bufferOffset int
	buffer       []int
	limit        int
	level        int
	end          int
}

func NewSliceReader(pool *IntBlockPool) *SliceReader {
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

	firstSize := LEVEL_SIZE_ARRAY[0]
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
	s.level = NEXT_LEVEL_ARRAY[s.level]
	newSize := LEVEL_SIZE_ARRAY[s.level]

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
