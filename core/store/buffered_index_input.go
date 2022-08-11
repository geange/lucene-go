package store

import (
	"bytes"
)

// BufferedIndexInput Base implementation class for buffered IndexInput.
type BufferedIndexInput interface {
	IndexInput
	//RandomAccessInput

	// ReadInternal Expert: implements buffer refill. Reads bytes from the current position in the input.
	// Params: b – the buffer to read bytes into
	//ReadInternal(buf *bytes.Buffer) error

	// SeekInternal Expert: implements seek. Sets current position in this file, where the next
	// readInternal(ByteBuffer) will occur.
	// See Also: readInternal(ByteBuffer)
	//SeekInternal(pos int) error

	BufferedIndexInputExt
}

type BufferedIndexInputExt interface {
	// Change the buffer size used by this IndexInput
	SetBufferSize(newSize int)
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

type BufferedIndexInputImp struct {
	bufferSize  int
	bufferStart int
	buffer      *bytes.Buffer
}
