package store

import (
	"bytes"
	"errors"
	"io"
)

// BufferedIndexInput Base implementation class for buffered IndexInput.
type BufferedIndexInput interface {
	IndexInput
	//RandomAccessInput

	BufferedIndexInputInner

	SetBufferSize(newSize int)
}

type BufferedIndexInputInner interface {
	// ReadInternal Expert: implements buffer refill. Reads bytes from the current pos in the input.
	// Params: b – the buffer to read bytes into
	ReadInternal(buf *bytes.Buffer, size int) error

	// SeekInternal Expert: implements seek. Sets current pos in this file, where the next
	// readInternal(ByteBuffer) will occur.
	// See Also: readInternal(ByteBuffer)
	SeekInternal(pos int) error
}

const (
	// BUFFER_SIZE Default buffer size set to 1024.
	BUFFER_SIZE = 1024

	// MIN_BUFFER_SIZE Minimum buffer size allowed
	MIN_BUFFER_SIZE = 8

	// MERGE_BUFFER_SIZE The normal read buffer size defaults to 1024, but
	// increasing this during merging seems to yield
	// performance gains.  However we don't want to increase
	// it too much because there are quite a few
	// BufferedIndexInputs created during merging.  See
	// LUCENE-888 for details.
	// A buffer size for merges set to 4096.
	MERGE_BUFFER_SIZE = 4096
)

type BufferedIndexInputBase struct {
	*IndexInputBase

	bufferSize     int
	bufferStart    int // pos in file of buffer
	bufferPosition int // pos in buffer
	buffer         *bytes.Buffer

	buffInner BufferedIndexInputInner
}

func NewBufferedIndexInputBase(bInput BufferedIndexInput) *BufferedIndexInputBase {
	input := &BufferedIndexInputBase{
		bufferSize: BUFFER_SIZE,
		buffInner:  bInput,
	}

	input.IndexInputBase = NewIndexInputBase(bInput)
	return input
}

func (r *BufferedIndexInputBase) Clone(bInput BufferedIndexInput) *BufferedIndexInputBase {
	var buffer *bytes.Buffer
	if r.buffer != nil {
		buffer = bytes.NewBuffer(r.buffer.Bytes())
	} else {
		buffer = new(bytes.Buffer)
	}

	return &BufferedIndexInputBase{
		IndexInputBase: r.IndexInputBase.Clone(bInput),
		bufferSize:     r.bufferSize,
		bufferStart:    r.bufferStart,
		bufferPosition: r.bufferPosition,
		buffer:         buffer,
		buffInner:      bInput,
	}
}

func (r *BufferedIndexInputBase) ReadByte() (byte, error) {
	if r.buffer == nil {
		if err := r.refill(); err != nil {
			return 0, err
		}
	}

	r.bufferPosition++
	b, err := r.buffer.ReadByte()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return 0, err
		}
		if err := r.refill(); err != nil {
			return 0, err
		}
		return r.buffer.ReadByte()
	}
	return b, nil
}

func (r *BufferedIndexInputBase) Read(b []byte) (n int, err error) {
	if r.buffer == nil {
		if err := r.refill(); err != nil {
			return 0, err
		}
	}

	available := r.buffer.Len()
	size := len(b)

	if size <= available {
		r.bufferSize += size
		return r.buffer.Read(b)
	}

	// the buffer does not have enough data. First serve all we've got.
	if _, err := r.buffer.Read(b[:available]); err != nil {
		return 0, err
	}
	r.bufferStart += available
	r.buffer.Reset()
	r.bufferPosition = 0

	leftSize := len(b) - available
	newBuffer := bytes.NewBuffer(b[available:])
	newBuffer.Reset()
	if err := r.buffInner.ReadInternal(newBuffer, leftSize); err != nil {
		return 0, err
	}
	r.bufferStart += newBuffer.Len()
	return newBuffer.Len() + available, nil
}

func (r *BufferedIndexInputBase) GetFilePointer() int64 {
	return int64(r.bufferStart + r.bufferPosition)
}

// SetBufferSize Change the buffer size used by this IndexInput
func (r *BufferedIndexInputBase) SetBufferSize(newSize int) {
	if newSize != r.bufferSize && newSize >= MIN_BUFFER_SIZE {
		r.bufferSize = newSize
		if r.buffer != nil {
			newBuffer := bytes.NewBuffer(make([]byte, 0, newSize))
			bs := r.buffer.Bytes()
			if len(bs) > newSize {
				bs = bs[:newSize]
			}
			newBuffer.Write(bs)
			r.buffer = newBuffer
			r.bufferPosition = 0
		}
	}
}

func (r *BufferedIndexInputBase) refill() error {
	start := r.bufferStart + r.bufferPosition
	if r.buffer == nil {
		r.buffer = bytes.NewBuffer(make([]byte, r.bufferSize))
		if err := r.buffInner.SeekInternal(r.bufferStart); err != nil {
			return err
		}
	}
	r.buffer.Reset()
	r.bufferStart = start

	return r.buffInner.ReadInternal(r.buffer, r.bufferSize)
}

// Change the buffer size used by this IndexInput
