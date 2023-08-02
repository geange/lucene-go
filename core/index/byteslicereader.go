package index

import (
	"encoding/binary"
	"errors"
	"slices"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/bytesref"
)

var _ store.DataInput = &ByteSliceReader{}

// ByteSliceReader IndexInput that knows how to read the byte slices written
// by Posting and PostingVector.  We read the bytes in
// each slice until we hit the end of that slice at which
// point we read the forwarding address of the next slice
// and then jump to it.
type ByteSliceReader struct {
	*store.BaseDataInput

	pool         *bytesref.BlockPool
	bufferUpto   int
	buffer       []byte
	upto         int
	limit        int
	level        int
	bufferOffset int
	endIndex     int
}

func (b *ByteSliceReader) Clone() store.CloneReader {
	input := &ByteSliceReader{
		pool:         b.pool,
		bufferUpto:   b.bufferUpto,
		buffer:       slices.Clone(b.buffer),
		upto:         b.upto,
		limit:        b.limit,
		level:        b.level,
		bufferOffset: b.bufferOffset,
		endIndex:     b.endIndex,
	}
	input.BaseDataInput = store.NewBaseDataInput(input)
	return input
}

func NewByteSliceReader() *ByteSliceReader {
	return &ByteSliceReader{}
}

func (b *ByteSliceReader) init(pool *bytesref.BlockPool, startIndex, endIndex int) error {
	b.pool = pool
	b.endIndex = endIndex

	b.level = 0
	b.bufferUpto = startIndex / bytesref.BYTE_BLOCK_SIZE
	b.bufferOffset = b.bufferUpto * bytesref.BYTE_BLOCK_SIZE
	b.buffer = pool.GetBytes(uint32(b.bufferUpto))
	b.upto = startIndex & bytesref.BYTE_BLOCK_MASK

	firstSize := bytesref.LEVEL_SIZE_ARRAY[0]
	if startIndex+firstSize >= endIndex {
		// There is only this one slice to read
		b.limit = endIndex & bytesref.BYTE_BLOCK_MASK
	} else {
		b.limit = b.upto + firstSize - 4
	}

	return nil
}

func (b *ByteSliceReader) Read(bs []byte) (n int, err error) {
	offset := 0
	size := len(bs)
	for size > 0 {
		numLeft := b.limit - b.upto
		if numLeft < size {
			copy(bs[offset:], b.buffer[b.upto:b.upto+numLeft])
			offset += numLeft
			size -= numLeft
			b.nextSlice()
		} else {
			copy(bs[offset:], b.buffer[b.upto:])
			b.upto += size
			return size, nil
		}
	}
	return 0, errors.New("size of bs is zero")
}

func (b *ByteSliceReader) nextSlice() {
	// Skip to our next slice
	nextIndex := binary.BigEndian.Uint32(b.buffer[b.limit:])
	b.level = bytesref.NEXT_LEVEL_ARRAY[b.level]
	newSize := bytesref.LEVEL_SIZE_ARRAY[b.level]

	b.bufferUpto = int(nextIndex / bytesref.BYTE_BLOCK_SIZE)
	b.bufferOffset = b.bufferUpto * bytesref.BYTE_BLOCK_SIZE
	b.buffer = b.pool.Get(b.bufferUpto)
	b.upto = int(nextIndex & bytesref.BYTE_BLOCK_MASK)

	if int(nextIndex)+newSize >= b.endIndex {
		// We are advancing to the final slice
		// assert endIndex - nextIndex > 0;
		b.limit = b.endIndex - b.bufferOffset
	} else {
		// This is not the final slice (subtract 4 for the
		// forwarding address at the end of this new slice)
		b.limit = b.upto + newSize - 4
	}
}
