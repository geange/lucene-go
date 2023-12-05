package store

import (
	"hash"
	"hash/crc32"
)

var _ IndexOutput = &RAMOutputStream{}

// RAMOutputStream A memory-resident IndexOutput implementation.
// Deprecated This class uses inefficient synchronization and is discouraged in favor of MMapDirectory.
// It will be removed in future versions of Lucene.
type RAMOutputStream struct {
	//*IndexOutputImp
	*IndexOutputBase

	bufferSize         int
	file               *RAMFile
	currentBuffer      []byte
	currentBufferIndex int
	bufferPosition     int
	bufferStart        int
	bufferLength       int
	checksum           bool
	crc                hash.Hash32
}

func NewRAMOutputStream() *RAMOutputStream {
	return NewRAMOutputStreamV1("noname", NewRAMFile(), false)
}

func (r *RAMOutputStream) WriteByte(b byte) error {
	if r.bufferPosition == r.bufferLength {
		r.currentBufferIndex++
		r.switchCurrentBuffer()
	}
	if r.crc != nil {
		if _, err := r.crc.Write([]byte{b}); err != nil {
			return err
		}
	}
	r.currentBuffer[r.bufferPosition] = b
	r.bufferPosition++
	return nil
}

func (r *RAMOutputStream) Write(b []byte) (int, error) {
	if r.crc != nil {
		if _, err := r.crc.Write(b); err != nil {
			return 0, err
		}
	}

	offset := 0
	size := len(b)

	for size > 0 {
		if r.bufferPosition == r.bufferLength {
			r.currentBufferIndex++
			r.switchCurrentBuffer()
		}

		remainInBuffer := len(r.currentBuffer) - r.bufferPosition
		bytesToCopy := size
		if size >= remainInBuffer {
			bytesToCopy = remainInBuffer
		}
		copy(r.currentBuffer[r.bufferPosition:], b[offset:offset+bytesToCopy])
		offset += bytesToCopy
		size -= bytesToCopy
		r.bufferPosition += bytesToCopy
	}

	return size, nil
}

func NewRAMOutputStreamV1(name string, f *RAMFile, checksum bool) *RAMOutputStream {
	output := &RAMOutputStream{

		bufferSize:         1024,
		file:               f,
		currentBufferIndex: -1,
		checksum:           checksum,
	}

	output.IndexOutputBase = NewIndexOutputBase(name, output)

	if checksum {
		output.crc = crc32.NewIEEE()
	}
	return output
}

// WriteTo Copy the current contents of this buffer to the provided DataOutput.
func (r *RAMOutputStream) WriteTo(out DataOutput) error {
	err := r.flush()
	if err != nil {
		return err
	}

	end := int(r.file.length)
	pos, buffer := 0, 0

	for pos < end {
		length := r.bufferSize
		nextPos := pos + length
		if nextPos > end {
			length = end - pos
		}
		_, err := out.Write(r.file.getBuffer(buffer)[:length])
		if err != nil {
			return err
		}
		buffer++
		pos = nextPos
	}
	return nil
}

// WriteToV1 Copy the current contents of this buffer to output byte array
func (r *RAMOutputStream) WriteToV1(bytes []byte) error {
	err := r.flush()
	if err != nil {
		return err
	}

	end := int(r.file.length)
	pos, buffer, bytesUpto := 0, 0, 0

	for pos < end {
		length := r.bufferSize
		nextPos := pos + length
		if nextPos > end {
			length = end - pos
		}

		src := r.file.getBuffer(buffer)[:length]
		copy(bytes[bytesUpto:], src)
		if err != nil {
			return err
		}
		buffer++
		bytesUpto += length
		pos = nextPos
	}
	return nil
}

func (r *RAMOutputStream) Close() error {
	return r.flush()
}

func (r *RAMOutputStream) switchCurrentBuffer() {
	if r.currentBufferIndex == r.file.numBuffers() {
		r.currentBuffer = r.file.addBuffer(r.bufferSize)
	} else {
		r.currentBuffer = r.file.getBuffer(r.currentBufferIndex)
	}

	r.bufferPosition = 0
	r.bufferStart = r.bufferSize * r.currentBufferIndex
	r.bufferLength = len(r.currentBuffer)
}

func (r *RAMOutputStream) setFileLength() {
	pointer := r.bufferStart + r.bufferPosition
	if pointer > int(r.file.length) {
		r.file.setLength(int64(pointer))
	}
}

func (r *RAMOutputStream) flush() error {
	r.setFileLength()
	return nil
}

func (r *RAMOutputStream) GetFilePointer() int64 {
	if r.currentBufferIndex < 0 {
		return 0
	}
	return int64(r.bufferStart + r.bufferPosition)
}

func (r *RAMOutputStream) GetChecksum() (uint32, error) {
	return r.crc.Sum32(), nil
}

// Reset Resets this to an empty file.
func (r *RAMOutputStream) Reset() {
	r.currentBuffer = nil
	r.currentBufferIndex = -1
	r.bufferPosition = 0
	r.bufferStart = 0
	r.bufferLength = 0
	r.file.setLength(0)
	if r.crc != nil {
		r.crc.Reset()
	}
}
