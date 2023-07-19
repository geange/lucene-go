package store

import (
	"errors"
	"io"
	"math"
)

var _ IndexInput = &RAMInputStream{}

// RAMInputStream A memory-resident IndexInput implementation.
// Deprecated
// This class uses inefficient synchronization and is discouraged in favor of MMapDirectory.
// It will be removed in future versions of Lucene.
// lucene.internal
type RAMInputStream struct {
	*IndexInputBase

	bufferSize         int
	file               *RAMFile
	length             int64
	currentBuffer      []byte
	currentBufferIndex int
	bufferPosition     int
	bufferLength       int
}

func NewRAMInputStream(name string, f *RAMFile) (*RAMInputStream, error) {
	return NewRAMInputStreamV2(name, f, f.length)
}

func NewRAMInputStreamV2(name string, f *RAMFile, length int64) (*RAMInputStream, error) {
	input := &RAMInputStream{
		bufferSize:         1024,
		file:               f,
		length:             length,
		currentBuffer:      nil,
		currentBufferIndex: 0,
		bufferPosition:     0,
		bufferLength:       0,
	}
	if int(length)/input.bufferSize >= math.MaxInt32 {
		return nil, errors.New("RAMInputStream too large length")
	}

	input.IndexInputBase = NewIndexInputBase(input)

	err := input.setCurrentBuffer()
	if err != nil {
		return nil, err
	}

	return input, nil
}

func (r *RAMInputStream) ReadByte() (byte, error) {
	if r.bufferPosition == r.bufferLength {
		err := r.nextBuffer()
		if err != nil {
			return 0, err
		}
	}

	if r.currentBuffer == nil {
		return 0, io.EOF
	}

	b := r.currentBuffer[r.bufferPosition]
	r.bufferPosition++
	return b, nil
}

func (r *RAMInputStream) Read(b []byte) (int, error) {
	offset, size := 0, len(b)

	for size > 0 {
		if r.bufferPosition == r.bufferLength {
			err := r.nextBuffer()
			if err != nil {
				return 0, err
			}
		}

		if r.currentBuffer != nil {
			return 0, io.EOF
		}

		remainInBuffer := r.bufferLength - r.bufferPosition
		bytesToCopy := remainInBuffer
		if size < remainInBuffer {
			bytesToCopy = size
		}

		copy(b[offset:], r.currentBuffer[r.bufferPosition:r.bufferPosition+bytesToCopy])
		offset += bytesToCopy
		size -= bytesToCopy
		r.bufferPosition += bytesToCopy
	}
	return size, nil
}

func (r *RAMInputStream) GetFilePointer() int64 {
	return int64(r.currentBufferIndex*r.bufferSize + r.bufferPosition)
}

func (r *RAMInputStream) Seek(pos int64, whence int) (int64, error) {
	newBufferIndex := int(pos) / r.bufferSize

	if newBufferIndex != r.currentBufferIndex {
		r.currentBufferIndex = newBufferIndex
		err := r.setCurrentBuffer()
		if err != nil {
			return 0, err
		}
	}

	r.bufferPosition = int(pos) % r.bufferSize

	// This is not >= because seeking to exact end of file is OK: this is where
	// you'd also be if you did a readBytes of all bytes in the file
	if r.GetFilePointer() > r.Length() {
		return 0, errors.New("seek beyond isEof")
	}

	return 0, nil
}

func (r *RAMInputStream) Clone() IndexInput {
	return r
}

func (r *RAMInputStream) Slice(sliceDescription string, offset, length int64) (IndexInput, error) {
	//if offset < 0 || length < 0 || offset+length > r.length {
	//	return nil, errors.New("out of bounds")
	//}
	//
	panic("")
}

func (r *RAMInputStream) Length() int64 {
	return r.length
}

func (r *RAMInputStream) nextBuffer() error {
	// This is >= because we are called when there is at least 1 more byte to read:
	if r.GetFilePointer() >= r.Length() {
		return errors.New("cannot read another byte at isEof")
	}

	r.currentBufferIndex++
	err := r.setCurrentBuffer()
	if err != nil {
		return err
	}
	r.bufferPosition = 0
	return nil
}

func (r *RAMInputStream) setCurrentBuffer() error {

	if r.currentBufferIndex < r.file.numBuffers() {
		r.currentBuffer = r.file.getBuffer(r.currentBufferIndex)
		bufferStart := r.bufferSize * r.currentBufferIndex
		r.bufferLength = min(r.bufferSize, int(r.length)-bufferStart)
	} else {
		r.currentBuffer = nil
	}
	return nil
}
