package bytesutils

import (
	"encoding/binary"
	"io"
)

const (
	BlockShift = 15
	BlockSize  = 1 << BlockShift
	BlockMask  = BlockSize - 1
)

var (
	// ByteNextLevelArray An array holding the offset into the LEVEL_SIZE_ARRAY to quickly navigate to the next slice level.
	ByteNextLevelArray = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 9}

	// ByteLevelSizeArray An array holding the level sizes for byte slices.
	ByteLevelSizeArray = []int{5, 14, 20, 30, 40, 40, 80, 80, 120, 200}

	// ByteFirstLevelSize The first level size for new slices
	// See Also: NewSlice(int)
	ByteFirstLevelSize = 5
)

// BlockPool that Posting and PostingVector use to write byte streams into shared fixed-size
// byte[] arrays. The idea is to allocate slices of increasing lengths For example, the first slice is 5
// bytes, the next slice is 14, etc. We start by writing our bytes into the first 5 bytes. When we hit
// the end of the slice, we allocate the next slice and then write the address of the new slice into the
// last 4 bytes of the previous slice (the "forwarding address"). Each slice is filled with 0's initially,
// and we mark the end with a non-zero byte. This way the methods that are writing into the slice don't
// need to record its length and instead allocate a new slice once they hit a non-zero byte.
type BlockPool struct {
	buffers    [][]byte  // array of buffers currently used in the pool. Buffers are allocated if needed don't modify this outside of this class.
	bufferUpto int       // index into the buffers array pointing to the current buffer used as the head Which buffer we are upto
	byteUpto   int       // Where we are in head buffer
	buffer     []byte    // Current head buffer
	byteOffset int       // Current head offset
	allocator  Allocator //
}

func NewBlockPool(allocator Allocator) *BlockPool {
	return &BlockPool{
		buffers:    make([][]byte, 0, 10),
		bufferUpto: -1,
		byteUpto:   BlockSize,
		byteOffset: -BlockSize,
		allocator:  allocator,
	}
}

// Reset Expert: Resets the pool to its initial state reusing the first buffer. Calling nextBuffer() is not
// needed after reset.
// zeroFillBuffers: if true the buffers are filled with 0. This should be set to true if this pool is used with slices.
// reuseFirst: if true the first buffer will be reused and calling nextBuffer() is not needed after
//
//	reset iff the block pool was used before ie. nextBuffer() was called before.
func (r *BlockPool) Reset(zeroFillBuffers, reuseFirst bool) {
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
			r.byteOffset = 0
			r.buffer = r.buffers[0]
		} else {
			r.bufferUpto = -1
			r.byteUpto = BlockSize
			r.byteOffset = -BlockSize
			r.buffer = nil
		}
	}
}

// NextBuffer Advances the pool to its next buffer. This method should be called once after the constructor to
// initialize the pool. In contrast to the constructor a reset() call will advance the pool to its first buffer
// immediately.
func (r *BlockPool) NextBuffer() {
	r.buffers = append(r.buffers, r.allocator.GetByteBlock())
	r.buffer = r.buffers[r.bufferUpto+1]
	r.bufferUpto++
	r.byteUpto = 0
	r.byteOffset += BlockSize
}

// NewSlice Allocates a new slice with the given size.
// See Also: FIRST_LEVEL_SIZE
func (r *BlockPool) NewSlice(size int) int {
	if r.byteUpto > BlockSize-size {
		r.NextBuffer()
	}

	upto := r.byteUpto
	r.byteUpto += size
	r.buffer[r.byteUpto-1] = 16
	return upto
}

// AllocSlice Creates a new byte slice with the given starting size and returns the slices offset in the pool.
func (r *BlockPool) AllocSlice(slice []byte, upto int) int {
	level := slice[upto] & 15
	newLevel := ByteNextLevelArray[level]
	newSize := ByteLevelSizeArray[newLevel]

	// Maybe allocate another block
	if r.byteUpto > BlockSize-newSize {
		r.NextBuffer()
	}

	newUpto := r.byteUpto
	offset := newUpto + r.byteOffset
	r.byteUpto += newSize

	// Copy forward the past 3 bytes (which we are about
	// to overwrite with the forwarding address):
	r.buffer[newUpto] = slice[upto-3]
	r.buffer[newUpto+1] = slice[upto-2]
	r.buffer[newUpto+2] = slice[upto-1]

	// Write forwarding address at end of last slice:
	binary.BigEndian.PutUint32(slice[upto-3:], uint32(offset))
	//slice[upto-3] = byte(offset >> 24)
	//slice[upto-2] = byte(offset >> 16)
	//slice[upto-1] = byte(offset >> 8)
	//slice[upto] = byte(offset)

	// Write new level:
	r.buffer[r.byteUpto-1] = byte(16 | newLevel)

	return newUpto + 3
}

// SetBytesRefV1 Fill the provided BytesRef with the bytes at the specified offset/length slice. This will
// avoid copying the bytes, if the slice fits into a single block; otherwise, it uses the provided BytesRefBuilder
// to copy bytes over.
func (r *BlockPool) SetBytesRefV1(builder *BytesRefBuilder, result []byte, offset, length int) {
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

func (r *BlockPool) GetAddress(offset int) ([]byte, error) {
	bufferIndex := offset >> BlockShift
	pos := offset & BlockMask
	values := r.buffers[bufferIndex]

	size, n := binary.Uvarint(values[pos:])
	if size == 0 {
		return nil, io.EOF
	}

	from := pos + n
	to := from + int(size)
	return values[from:to], nil
}

func (r *BlockPool) Get(index int) []byte {
	return r.buffers[index]
}

func (r *BlockPool) ByteUpto() int {
	return r.byteUpto
}

func (r *BlockPool) GetBytes(textStart int) []byte {
	bytes := r.buffers[textStart>>BlockShift]

	pos := textStart & BlockMask

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

func (r *BlockPool) get(textStart int) []byte {
	bytes := r.buffers[textStart>>BlockShift]

	pos := textStart & BlockMask

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
func (r *BlockPool) ReadBytes(offset int, bytes []byte, bytesOffset, bytesLength int) {
	bytesLeft := bytesLength
	bufferIndex := offset >> BlockShift
	pos := offset * BlockMask

	for bytesLeft > 0 {
		buffer := r.buffers[bufferIndex]
		bufferIndex++
		chunk := min(bytesLeft, BlockSize-pos)
		copy(bytes[bytesOffset:], buffer[pos:pos+chunk])
		bytesOffset += chunk
		bytesLeft -= chunk
		pos = 0
	}
}

// SetRawBytesRef Set the given BytesRef so that its content is equal to the ref.length bytes starting at offset. Most of the time this method will set pointers to internal data-structures. However, in case a value crosses a boundary, a fresh copy will be returned. On the contrary to setBytesRef(BytesRef, int), this does not expect the length to be encoded with the data.
func (r *BlockPool) SetRawBytesRef(ref *BytesRef, offset int) {
	bufferIndex := offset >> BlockShift
	pos := offset & BlockMask
	if pos+ref.length <= BlockSize {
		ref.bs = r.buffers[bufferIndex]
		ref.offset = pos
	} else {
		ref.bs = make([]byte, ref.length)
		ref.offset = 0
		r.ReadBytes(offset, ref.bs, 0, ref.length)
	}
}

// Append Appends the bytes in the provided BytesRef at the current position.
func (r *BlockPool) Append(bytes []byte) {
	bytesLeft := len(bytes)
	offset := 0

	for bytesLeft > 0 {
		bufferLeft := BlockSize - r.byteUpto
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

func (r *BlockPool) Current() []byte {
	return r.buffer
}

func (r *BlockPool) ByteOffset() int {
	return r.byteOffset
}
