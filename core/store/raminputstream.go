package store

import (
	"io"
	"slices"
)

var _ IndexInput = &RAMInputStream{}

type RAMInputStream struct {
	*BaseIndexInput

	file               *RAMFile
	length             int
	currentBuffer      []byte
	currentBufferIndex int
	bufferLength       int
	bufferPosition     int
}

func NewRAMInputStream(name string, file *RAMFile, length int) (*RAMInputStream, error) {
	stream := &RAMInputStream{
		file:   file,
		length: length,
	}

	stream.BaseIndexInput = NewBaseIndexInput(stream)
	if err := stream.setCurrentBuffer(); err != nil {
		return nil, err
	}
	return stream, nil
}

func (s *RAMInputStream) Read(p []byte) (n int, err error) {
	size := len(p)
	dstIdx := 0

	for size > 0 {
		if s.bufferPosition == s.bufferLength {
			if err := s.nextBuffer(); err != nil {
				return 0, err
			}
		}

		buffSize := s.bufferLength - s.bufferPosition
		if buffSize > size {
			copy(p[dstIdx:], s.currentBuffer[s.bufferPosition:])
			break
		}

		size -= buffSize
		dstIdx += buffSize
	}
	return len(p), nil
}

func (s *RAMInputStream) nextBuffer() error {
	buff, ok := s.file.GetBuffer(s.currentBufferIndex)
	if !ok {
		return io.EOF
	}
	s.currentBuffer = buff
	s.bufferPosition = 0
	s.bufferLength = len(buff)
	return nil
}

func (s *RAMInputStream) Clone() CloneReader {
	stream := &RAMInputStream{
		file:               s.file.Clone(),
		currentBuffer:      slices.Clone(s.currentBuffer),
		currentBufferIndex: s.currentBufferIndex,
		bufferLength:       s.bufferLength,
		bufferPosition:     s.bufferPosition,
	}
	stream.BaseIndexInput = NewBaseIndexInput(stream)
	return stream
}

func (s *RAMInputStream) Seek(offset int64, whence int) (int64, error) {

	newBufferIndex := int(offset / RAM_BUFFER_SIZE)

	if newBufferIndex != s.currentBufferIndex {
		s.currentBufferIndex = newBufferIndex
		if err := s.setCurrentBuffer(); err != nil {
			return 0, err
		}
	}

	s.bufferPosition = int(offset % RAM_BUFFER_SIZE)

	// This is not >= because seeking to exact end of file is OK: this is where
	// you'd also be if you did a readBytes of all bytes in the file
	if s.GetFilePointer() > s.Length() {
		return 0, io.ErrUnexpectedEOF
	}
	return offset, nil
}

func (s *RAMInputStream) GetFilePointer() int64 {
	return int64(RAM_BUFFER_SIZE*s.currentBufferIndex + s.bufferPosition)
}

func (s *RAMInputStream) Slice(sliceDescription string, offset, length int64) (IndexInput, error) {
	//TODO implement me
	panic("implement me")
}

func (s *RAMInputStream) Length() int64 {
	return s.file.GetLength()
}

func (s *RAMInputStream) setCurrentBuffer() error {
	if s.currentBufferIndex < s.file.NumBuffers() {
		buffer, ok := s.file.GetBuffer(s.currentBufferIndex)
		if !ok {
			return io.ErrUnexpectedEOF
		}
		s.currentBuffer = buffer
		bufferStart := RAM_BUFFER_SIZE * s.currentBufferIndex
		s.bufferLength = min(RAM_BUFFER_SIZE, s.length-bufferStart)
	} else {
		s.currentBuffer = nil
	}
	return nil
}
