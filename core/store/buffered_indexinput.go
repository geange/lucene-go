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

	// ReadInternal Expert: implements buffer refill. Reads bytes from the current position in the input.
	// Params: b – the buffer to read bytes into
	ReadInternal(buf *bytes.Buffer, size int) error

	// SeekInternal Expert: implements seek. Sets current position in this file, where the next
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

type BufferedIndexInputDefaultConfig struct {
	IndexInputDefaultConfig

	ReadInternal func(buf *bytes.Buffer, size int) error
	SeekInternal func(pos int) error
}

type BufferedIndexInputDefault struct {
	*IndexInputDefault

	bufferSize     int
	bufferStart    int // position in file of buffer
	bufferPosition int // position in buffer
	buffer         *bytes.Buffer

	readInternal func(buf *bytes.Buffer, size int) error
	seekInternal func(pos int) error
}

func NewBufferedIndexInputDefault(cfg *BufferedIndexInputDefaultConfig) *BufferedIndexInputDefault {
	input := &BufferedIndexInputDefault{
		bufferSize:   BUFFER_SIZE,
		readInternal: cfg.ReadInternal,
		seekInternal: cfg.SeekInternal,
	}

	input.IndexInputDefault = NewIndexInputDefault(&IndexInputDefaultConfig{
		DataInputDefaultConfig: DataInputDefaultConfig{
			ReadByte: input.ReadByte,
			Read:     input.Read,
		},
		Close:          cfg.Close,
		GetFilePointer: cfg.GetFilePointer,
		Seek:           cfg.Seek,
		Slice:          cfg.Slice,
		Length:         cfg.Length,
	})
	return input
}

func (r *BufferedIndexInputDefault) Clone(cfg *BufferedIndexInputDefaultConfig) *BufferedIndexInputDefault {
	return &BufferedIndexInputDefault{
		IndexInputDefault: r.IndexInputDefault.Clone(&cfg.IndexInputDefaultConfig),
		bufferSize:        r.bufferSize,
		bufferStart:       r.bufferStart,
		bufferPosition:    r.bufferPosition,
		buffer:            bytes.NewBuffer(r.buffer.Bytes()),
		readInternal:      cfg.ReadInternal,
		seekInternal:      cfg.SeekInternal,
	}
}

func (r *BufferedIndexInputDefault) ReadByte() (byte, error) {
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

func (r *BufferedIndexInputDefault) Read(b []byte) (n int, err error) {
	available := r.buffer.Len()
	size := len(b)

	if size <= available {
		r.bufferSize += size
		return r.buffer.Read(b)
	}

	if available <= 0 {
		return 0, nil
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
	if err := r.readInternal(newBuffer, leftSize); err != nil {
		return 0, err
	}
	r.bufferStart += newBuffer.Len()
	return newBuffer.Len() + available, nil
}

func (r *BufferedIndexInputDefault) GetFilePointer() int64 {
	return int64(r.bufferStart + r.bufferPosition)
}

// SetBufferSize Change the buffer size used by this IndexInput
func (r *BufferedIndexInputDefault) SetBufferSize(newSize int) {
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

func (r *BufferedIndexInputDefault) refill() error {
	start := r.bufferStart + r.bufferPosition
	if r.buffer == nil {
		r.buffer = bytes.NewBuffer(make([]byte, r.bufferSize))
		if err := r.seekInternal(r.bufferStart); err != nil {
			return err
		}
	}
	r.buffer.Reset()
	r.bufferStart = start

	return r.readInternal(r.buffer, r.bufferSize)
}

// Change the buffer size used by this IndexInput
