package bytesref

import (
	"iter"
)

// Array
// A simple append only random-access BytesRef array that stores full copies of the
// appended []byte in a BlockPool.
// Note: This class is not Thread-Safe!
type Array struct {
	pool          *BlockPool
	offsets       []int
	lastElement   int
	currentOffset int
}

func NewArray(bytesUsed int64) *Array {
	array := newBytesRefArray()
	array.pool = NewBlockPool(newDirectAllocator(BYTE_BLOCK_SIZE))
	array.pool.NextBuffer()
	return array
}

func newBytesRefArray() *Array {
	return &Array{
		pool:          nil,
		offsets:       make([]int, 1),
		lastElement:   0,
		currentOffset: 0,
	}
}

func (r *Array) Append(bytes []byte) int {
	if r.lastElement >= len(r.offsets) {
		r.offsets = append(r.offsets, 0)
	}
	r.pool.Append(bytes)
	r.offsets[r.lastElement] = r.currentOffset
	r.lastElement++
	r.currentOffset += len(bytes)
	return r.lastElement - 1
}

func (r *Array) Clear() {
	r.lastElement = 0
	r.currentOffset = 0
	for i := range r.offsets {
		r.offsets[i] = 0
	}
	r.pool.Reset(false, true)
}

func (r *Array) Size() int {
	return r.lastElement
}

func (r *Array) Get(spare *Builder, index int) []byte {
	offset := r.offsets[index]
	var length int
	if index == r.lastElement-1 {
		length = r.currentOffset - offset
	} else {
		length = r.offsets[index+1] - offset
	}

	spare.Grow(length)
	spare.SetLength(length)
	r.pool.ReadBytes(offset, spare.Bytes(), 0, spare.Length())
	return spare.Get()
}

// Used only by sort below, to set a BytesRef with the specified slice, avoiding copying bytes in the common
// case when the slice is contained in a single block in the byte block pool.
//func (r *Array) setBytesRef(spare *Builder, result []byte, index int) {
//	offset := r.offsets[index]
//	length := 0
//	if index == r.lastElement-1 {
//		length = r.currentOffset - offset
//	} else {
//		length = r.offsets[index+1] - offset
//	}
//	r.pool.SetBytes(result[:length], offset)
//}

func (r *Array) Iterator() iter.Seq[[]byte] {
	size := r.Size()

	return func(yield func([]byte) bool) {
		for i := 0; i < size; i++ {
			offset := r.offsets[i]
			length := 0
			if i == r.lastElement-1 {
				length = r.currentOffset - offset
			} else {
				length = r.offsets[i+1] - offset
			}
			value, err := r.pool.GetBytesWithLength(offset, length)
			if err != nil {
				return
			}

			if !yield(value) {
				return
			}
		}
	}
}
