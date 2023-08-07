package util

import (
	"encoding/binary"
	"io"
)

const (
	BYTE_BLOCK_SHIFT = 15
	BYTE_BLOCK_SIZE  = 1 << BYTE_BLOCK_SHIFT
	BYTE_BLOCK_MASK  = BYTE_BLOCK_SIZE - 1
)

var (
	// BYTE_NEXT_LEVEL_ARRAY An array holding the offset into the LEVEL_SIZE_ARRAY to quickly navigate to the next slice level.
	BYTE_NEXT_LEVEL_ARRAY = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 9}

	// BYTE_LEVEL_SIZE_ARRAY An array holding the level sizes for byte slices.
	BYTE_LEVEL_SIZE_ARRAY = []int{5, 14, 20, 30, 40, 40, 80, 80, 120, 200}

	// BYTE_FIRST_LEVEL_SIZE The first level size for new slices
	// See Also: NewSlice(int)
	BYTE_FIRST_LEVEL_SIZE = 5
)

// ByteBlockPool Class that Posting and PostingVector use to write byte streams into shared fixed-size
// byte[] arrays. The idea is to allocate slices of increasing lengths For example, the first slice is 5
// bytes, the next slice is 14, etc. We start by writing our bytes into the first 5 bytes. When we hit
// the end of the slice, we allocate the next slice and then write the address of the new slice into the
// last 4 bytes of the previous slice (the "forwarding address"). Each slice is filled with 0's initially,
// and we mark the end with a non-zero byte. This way the methods that are writing into the slice don't
// need to record its length and instead allocate a new slice once they hit a non-zero byte.
type ByteBlockPool struct {
	// array of buffers currently used in the pool. Buffers are allocated if needed don't modify this
	// outside of this class.
	buffers [][]byte

	// index into the buffers array pointing to the current buffer used as the head
	bufferUpto int // Which buffer we are upto

	// Where we are in head buffer
	byteUpto int

	// Current head buffer
	buffer []byte

	// Current head offset
	ByteOffset int

	allocator BytesAllocator

	//NEXT_LEVEL_ARRAY []int

	//LEVEL_SIZE_ARRAY []int

	// The first level size for new slices
	// See Also: NewSlice(int)
	//FIRST_LEVEL_SIZE int
}

func NewByteBlockPool(allocator BytesAllocator) *ByteBlockPool {
	return &ByteBlockPool{
		buffers:    make([][]byte, 0, 10),
		bufferUpto: -1,
		byteUpto:   BYTE_BLOCK_SIZE,
		ByteOffset: -BYTE_BLOCK_SIZE,
		allocator:  allocator,
		//NEXT_LEVEL_ARRAY: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 9},
		//LEVEL_SIZE_ARRAY: []int{5, 14, 20, 30, 40, 40, 80, 80, 120, 200},
		//FIRST_LEVEL_SIZE: 5,
	}
}

// Reset Expert: Resets the pool to its initial state reusing the first buffer. Calling nextBuffer() is not
// needed after reset.
// Params: 	zeroFillBuffers – if true the buffers are filled with 0. This should be set to true if this pool is
//
//	used with slices.
//	reuseFirst – if true the first buffer will be reused and calling nextBuffer() is not needed after
//	reset iff the block pool was used before ie. nextBuffer() was called before.
func (r *ByteBlockPool) Reset(zeroFillBuffers, reuseFirst bool) {
	if r.bufferUpto != -1 {
		// We allocated at least one buffer
		if zeroFillBuffers {
			for idx := range r.buffers {
				// Fully zero fill buffers that we fully used
				for k := range r.buffers[idx] {
					r.buffers[idx][k] = 0
				}
			}
			// Partial zero fill the final buffer
		}

		if r.bufferUpto > 0 || !reuseFirst {
			offset := 0
			if reuseFirst {
				offset = 1
			}
			r.allocator.RecycleByteBlocks(r.buffers, offset, 1+r.bufferUpto)
			for idx := range r.buffers {
				if idx >= offset {
					r.buffers[idx] = nil
				}
			}
		}

		if reuseFirst {
			// Re-use the first buffer
			r.bufferUpto = 0
			r.byteUpto = 0
			r.ByteOffset = 0
			r.buffer = r.buffers[0]
		} else {
			r.bufferUpto = -1
			r.byteUpto = BYTE_BLOCK_SIZE
			r.ByteOffset = -BYTE_BLOCK_SIZE
			r.buffer = nil
		}
	}
}

// NextBuffer Advances the pool to its next buffer. This method should be called once after the constructor to
// initialize the pool. In contrast to the constructor a reset() call will advance the pool to its first buffer
// immediately.
func (r *ByteBlockPool) NextBuffer() {
	r.buffers = append(r.buffers, r.allocator.GetByteBlock())
	r.buffer = r.buffers[r.bufferUpto+1]
	r.bufferUpto++
	r.byteUpto = 0
	r.ByteOffset += BYTE_BLOCK_SIZE
}

// NewSlice Allocates a new slice with the given size.
// See Also: FIRST_LEVEL_SIZE
func (r *ByteBlockPool) NewSlice(size int) int {
	if r.byteUpto > INT_BLOCK_SIZE-size {
		r.NextBuffer()
	}

	upto := r.byteUpto
	r.byteUpto += size
	r.buffer[r.byteUpto-1] = 16
	return upto
}

// AllocSlice Creates a new byte slice with the given starting size and returns the slices offset in the pool.
func (r *ByteBlockPool) AllocSlice(slice []byte, upto int) int {
	level := slice[upto] & 15
	newLevel := BYTE_NEXT_LEVEL_ARRAY[level]
	newSize := BYTE_LEVEL_SIZE_ARRAY[newLevel]

	// Maybe allocate another block
	if r.byteUpto > BYTE_BLOCK_SIZE-newSize {
		r.NextBuffer()
	}

	newUpto := r.byteUpto
	offset := newUpto + r.ByteOffset
	r.byteUpto += newSize

	// Copy forward the past 3 bytes (which we are about
	// to overwrite with the forwarding address):
	r.buffer[newUpto] = slice[upto-3]
	r.buffer[newUpto+1] = slice[upto-2]
	r.buffer[newUpto+2] = slice[upto-1]

	// Write forwarding address at end of last slice:
	slice[upto-3] = byte(offset >> 24)
	slice[upto-2] = byte(offset >> 16)
	slice[upto-1] = byte(offset >> 8)
	slice[upto] = byte(offset)

	// Write new level:
	r.buffer[r.byteUpto-1] = byte(16 | newLevel)

	return newUpto + 3
}

// SetBytesRefV1 Fill the provided BytesRef with the bytes at the specified offset/length slice. This will
// avoid copying the bytes, if the slice fits into a single block; otherwise, it uses the provided BytesRefBuilder
// to copy bytes over.
func (r *ByteBlockPool) SetBytesRefV1(builder *BytesRefBuilder, result []byte, offset, length int) {
	//result.Len = length
	//
	//bufferIndex := offset >> BYTE_BLOCK_SHIFT
	//buffer := r.buffers[bufferIndex]
	//pos := offset & BYTE_BLOCK_MASK
	//if pos+length <= BYTE_BLOCK_SIZE {
	//	// common case where the slice lives in a single block: just reference the buffer directly without copying
	//	result.Bytes = buffer
	//	result.Offset = pos
	//} else {
	//	// uncommon case: the slice spans at least 2 blocks, so we must copy the bytes:
	//	builder.grow(length)
	//	result.Bytes = builder.Get()
	//	result.Offset = 0
	//	r.ReadBytes(offset, result.Bytes, 0, length)
	//}
}

func (r *ByteBlockPool) GetAddress(offset int) ([]byte, error) {
	bufferIndex := offset >> BYTE_BLOCK_SHIFT
	pos := offset & BYTE_BLOCK_MASK
	values := r.buffers[bufferIndex]

	size, n := binary.Uvarint(values[pos:])
	if size == 0 {
		return nil, io.EOF
	}

	from := pos + n
	to := from + int(size)
	return values[from:to], nil
}

func (r *ByteBlockPool) Get(index int) []byte {
	return r.buffers[index]
}

func (r *ByteBlockPool) ByteUpto() int {
	return r.byteUpto
}

func (r *ByteBlockPool) GetBytes(textStart int) []byte {
	bytes := r.buffers[textStart>>BYTE_BLOCK_SHIFT]

	pos := textStart & BYTE_BLOCK_MASK

	length, offset := 0, 0

	if (bytes[pos] & 0x80) == 0 {
		// length is 1 byte
		length = int(bytes[pos])
		offset = pos + 1
	} else {
		length = int((bytes[pos] & 0x7f) + (bytes[pos+1])<<7)
		offset = pos + 2
	}

	return bytes[offset : offset+length]
}

func (r *ByteBlockPool) get(textStart int) []byte {
	bytes := r.buffers[textStart>>BYTE_BLOCK_SHIFT]

	pos := textStart & BYTE_BLOCK_MASK

	if (bytes[pos] & 0x80) == 0 {
		// length is 1 byte
		size := int(bytes[pos])
		offset := pos + 1
		return bytes[offset : offset+size]
	} else {
		size := int((bytes[pos] & 0x7f) + (bytes[pos+1])<<7)
		offset := pos + 2
		return bytes[offset : offset+size]
	}
}

// ReadBytes Reads bytes out of the pool starting at the given offset with the given length into the given byte array at offset off.
// Note: this method allows to copy across block boundaries.
func (r *ByteBlockPool) ReadBytes(offset int, bytes []byte, bytesOffset, bytesLength int) {
	bytesLeft := bytesLength
	bufferIndex := offset >> BYTE_BLOCK_SHIFT
	pos := offset * BYTE_BLOCK_MASK

	for bytesLeft > 0 {
		buffer := r.buffers[bufferIndex]
		bufferIndex++
		chunk := min(bytesLeft, BYTE_BLOCK_SIZE-pos)
		copy(bytes[bytesOffset:], buffer[pos:pos+chunk])
		bytesOffset += chunk
		bytesLeft -= chunk
		pos = 0
	}
}

// SetRawBytesRef Set the given BytesRef so that its content is equal to the ref.length bytes starting at offset. Most of the time this method will set pointers to internal data-structures. However, in case a value crosses a boundary, a fresh copy will be returned. On the contrary to setBytesRef(BytesRef, int), this does not expect the length to be encoded with the data.
func (r *ByteBlockPool) SetRawBytesRef(ref *BytesRef, offset int) {
	bufferIndex := offset >> BYTE_BLOCK_SHIFT
	pos := offset & BYTE_BLOCK_MASK
	if pos+ref.Length <= BYTE_BLOCK_SIZE {
		ref.Bytes = r.buffers[bufferIndex]
		ref.Offset = pos
	} else {
		ref.Bytes = make([]byte, ref.Length)
		ref.Offset = 0
		r.ReadBytes(offset, ref.Bytes, 0, ref.Length)
	}
}

// Append Appends the bytes in the provided BytesRef at the current position.
func (r *ByteBlockPool) Append(bytes []byte) {
	bytesLeft := len(bytes)
	offset := 0

	for bytesLeft > 0 {
		bufferLeft := BYTE_BLOCK_SIZE - r.byteUpto
		if bytesLeft < bufferLeft {
			// fits within current buffer
			copy(r.buffer[r.byteUpto:], bytes[offset:offset+bytesLeft])
			r.byteUpto += bytesLeft
			break
		} else {
			// fill up this buffer and move to next one
			if bufferLeft > 0 {
				copy(r.buffer[r.byteUpto:], bytes[offset:offset+bufferLeft])
			}
			r.NextBuffer()
			bytesLeft -= bufferLeft
			offset += bufferLeft
		}
	}
}

func (r *ByteBlockPool) Current() []byte {
	return r.buffer
}

type BytesAllocator interface {
	RecycleByteBlocks(blocks [][]byte, start, end int)
	RecycleByteBlocksV1(blocks [][]byte)
	GetByteBlock() []byte
}

type BytesAllocatorExtra interface {
	RecycleByteBlocks(blocks [][]byte, start, end int)
}

var _ BytesAllocator = &BytesAllocatorDefault{}

type BytesAllocatorDefault struct {
	BlockSize           int
	FnRecycleByteBlocks func(blocks [][]byte, start, end int)
}

func (b *BytesAllocatorDefault) RecycleByteBlocks(blocks [][]byte, start, end int) {
	b.FnRecycleByteBlocks(blocks, start, end)
}

func (b *BytesAllocatorDefault) RecycleByteBlocksV1(blocks [][]byte) {
	b.FnRecycleByteBlocks(blocks, 0, len(blocks))
}

func (b *BytesAllocatorDefault) GetByteBlock() []byte {
	return make([]byte, b.BlockSize)
}

// BytesAllocatorImp Abstract class for allocating and freeing int blocks.
type BytesAllocatorImp struct {
	blockSize int
	BytesAllocatorExtra
}

func NewBytesAllocator(blockSize int, ext BytesAllocatorExtra) *BytesAllocatorImp {
	return &BytesAllocatorImp{
		blockSize:           blockSize,
		BytesAllocatorExtra: ext,
	}
}

func (r *BytesAllocatorImp) RecycleByteBlocksV1(blocks [][]byte) {
	r.RecycleByteBlocks(blocks, 0, len(blocks))
}

func (r *BytesAllocatorImp) RecycleByteBlocks(blocks [][]byte, start, end int) {
	r.RecycleByteBlocks(blocks, start, end)
}

func (r *BytesAllocatorImp) GetByteBlock() []byte {
	return make([]byte, r.blockSize)
}

type DirectBytesAllocator struct {
}

func NewDirectBytesAllocator() *DirectBytesAllocator {
	return &DirectBytesAllocator{}
}

func (d *DirectBytesAllocator) RecycleByteBlocks(blocks [][]byte, start, end int) {}
