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
	*IndexOutputImp

	bufferSize         int
	file               *RAMFile
	currentBuffer      []byte
	currentBufferIndex int
	bufferPosition     int
	bufferStart        int
	bufferLength       int
	crc                hash.Hash32
}

func NewRAMOutputStream() *RAMOutputStream {
	return NewRAMOutputStreamV1("noname", NewRAMFile(), false)
}

func NewRAMOutputStreamV1(name string, f *RAMFile, checksum bool) *RAMOutputStream {
	stream := &RAMOutputStream{
		bufferSize:         1024,
		file:               f,
		currentBufferIndex: -1,
	}

	stream.IndexOutputImp = NewIndexOutputImp(stream, name)

	if checksum {
		stream.crc = crc32.NewIEEE()
	}
	return stream
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
		err := out.WriteBytes(r.file.getBuffer(buffer)[:length])
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

func (r *RAMOutputStream) WriteBytes(b []byte) error {
	_, err := r.crc.Write(b)
	if err != nil {
		return err
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
		copy(r.copyBuffer[r.bufferPosition:], b[offset:offset+bytesToCopy])
		offset += bytesToCopy
		size -= bytesToCopy
		r.bufferPosition += bytesToCopy
	}

	return nil
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
