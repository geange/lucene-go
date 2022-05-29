package core

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
	bufferUpto int

	// Pointer to the current position in head buffer
	intUpto int

	// Current head buffer
	buffer []int

	// Current head offset
	intOffset int

	allocator *Allocator
}

func NewIntBlockPool() *IntBlockPool {
	return &IntBlockPool{
		buffers:    make([][]int, 10),
		bufferUpto: -1,
		intUpto:    INT_BLOCK_SIZE,
		buffer:     make([]int, 0),
		intOffset:  -INT_BLOCK_SIZE,
		allocator:  nil,
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
	i.buffer = i.buffers[i.bufferUpto]
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

// Allocator Abstract class for allocating and freeing int blocks.
type Allocator struct {
	blockSize int
	ext       AllocatorExt
}

func NewAllocator(blockSize int, ext AllocatorExt) *Allocator {
	return &Allocator{
		blockSize: blockSize,
		ext:       ext,
	}
}

func (a *Allocator) RecycleIntBlocks(blocks [][]int, start, end int) {
	a.ext.RecycleIntBlocks(blocks, start, end)
}

func (a *Allocator) GetIntBlock() []int {
	return make([]int, a.blockSize)
}

type AllocatorExt interface {
	RecycleIntBlocks(blocks [][]int, start, end int)
}
