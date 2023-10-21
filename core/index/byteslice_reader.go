package index

import (
	"encoding/binary"
	"errors"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/bytesutils"
)

var _ store.DataInput = &ByteSliceReader{}

// ByteSliceReader IndexInput that knows how to read the byte slices written
// by Posting and PostingVector.  We read the bytes in
// each slice until we hit the end of that slice at which
// point we read the forwarding address of the next slice
// and then jump to it.
type ByteSliceReader struct {
	store.Reader

	pool         *bytesutils.BlockPool
	bufferUpto   int
	buffer       []byte
	upto         int
	limit        int
	level        int
	bufferOffset int
	endIndex     int
}

func NewByteSliceReader() *ByteSliceReader {
	return &ByteSliceReader{}
}

func (b *ByteSliceReader) init(pool *bytesutils.BlockPool, startIndex, endIndex int) error {
	b.pool = pool
	b.endIndex = endIndex

	b.level = 0
	b.bufferUpto = startIndex / bytesutils.BlockSize
	b.bufferOffset = b.bufferUpto * bytesutils.BlockSize
	b.buffer = pool.GetBytes(b.bufferUpto)
	b.upto = startIndex & bytesutils.BlockMask

	firstSize := bytesutils.ByteLevelSizeArray[0]
	if startIndex+firstSize >= endIndex {
		// There is only this one slice to read
		b.limit = endIndex & bytesutils.BlockMask
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
	b.level = bytesutils.ByteNextLevelArray[b.level]
	newSize := bytesutils.ByteLevelSizeArray[b.level]

	b.bufferUpto = int(nextIndex / bytesutils.BlockSize)
	b.bufferOffset = b.bufferUpto * bytesutils.BlockSize
	b.buffer = b.pool.Get(b.bufferUpto)
	b.upto = int(nextIndex & bytesutils.BlockMask)

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
