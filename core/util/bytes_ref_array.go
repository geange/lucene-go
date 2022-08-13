package util

import (
	"go.uber.org/atomic"
	"io"
)

// BytesRefArray A simple append only random-access BytesRef array that stores full copies of the appended
// bytes in a ByteBlockPool. Note: This class is not Thread-Safe!
type BytesRefArray struct {
	pool          *ByteBlockPool
	offsets       []int
	lastElement   int
	currentOffset int
	bytesUsed     *atomic.Int64
}

func NewBytesRefArray(bytesUsed int) *BytesRefArray {
	array := newBytesRefArray()
	array.pool = NewByteBlockPool(NewBytesAllocator(BYTE_BLOCK_SIZE, NewDirectBytesAllocator()))
	array.pool.NextBuffer()
	// useless
	bytesUsed += 8
	return &BytesRefArray{bytesUsed: atomic.NewInt64(int64(bytesUsed))}
}

func newBytesRefArray() *BytesRefArray {
	return &BytesRefArray{
		pool:          nil,
		offsets:       make([]int, 1),
		lastElement:   0,
		currentOffset: 0,
		bytesUsed:     nil,
	}
}

func (r *BytesRefArray) Append(bytes []byte) int {
	if r.lastElement >= len(r.offsets) {
		r.offsets = append(r.offsets, 0)
	}
	r.pool.Append(bytes)
	r.offsets[r.lastElement] = r.currentOffset
	r.lastElement++
	r.currentOffset += len(bytes)
	return r.lastElement - 1
}

func (r *BytesRefArray) Clear() {
	r.lastElement = 0
	r.currentOffset = 0
	for i := range r.offsets {
		r.offsets[i] = 0
	}
	r.pool.Reset(false, true)
}

func (r *BytesRefArray) Size() int {
	return r.lastElement
}

func (r *BytesRefArray) Get(spare *BytesRefBuilder, index int) []byte {
	offset := r.offsets[index]
	length := func() int {
		if index == r.lastElement-1 {
			return r.currentOffset - offset
		}
		return r.offsets[index+1] - offset
	}()
	spare.Grow(length)
	spare.SetLength(length)
	r.pool.ReadBytes(offset, spare.Bytes(), 0, spare.Length())
	return spare.Get()
}

// Used only by sort below, to set a BytesRef with the specified slice, avoiding copying bytes in the common
// case when the slice is contained in a single block in the byte block pool.
func (r *BytesRefArray) setBytesRef(spare *BytesRefBuilder, result []byte, index int) {
	offset := r.offsets[index]
	length := 0
	if index == r.lastElement-1 {
		length = r.currentOffset - offset
	} else {
		length = r.offsets[index+1] - offset
	}
	r.pool.SetBytesRefV1(spare, result, offset, length)
}

func (r *BytesRefArray) Iterator(_ *BytesRef) BytesRefIterator {
	return &bytesRefIterator{
		size:  r.Size(),
		pos:   -1,
		ord:   0,
		spare: NewBytesRefBuilder(),
		ref:   []byte{},
		array: r,
	}
}

type bytesRefIterator struct {
	size  int
	pos   int
	ord   int
	spare *BytesRefBuilder
	ref   []byte
	array *BytesRefArray
}

func (b *bytesRefIterator) Next() ([]byte, error) {
	b.pos++
	if b.pos < b.size {
		b.ord = b.pos
		b.array.setBytesRef(b.spare, b.ref, b.ord)
		return b.ref, nil
	}
	return nil, io.EOF
}

type SortState struct {
	indices []int
}
