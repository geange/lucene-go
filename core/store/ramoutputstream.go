package store

import (
	"bytes"
	"hash"
	"hash/crc32"
)

const RAM_BUFFER_SIZE = 1024

var _ IndexOutput = &RAMOutputStream{}

type RAMOutputStream struct {
	*BaseIndexOutput

	bufferSize int
	file       *RAMFile
	buffer     *bytes.Buffer
	crc        hash.Hash32
	idx        int
}

type RAMOutputStreamOptionBuilder struct {
}

func NewRAMOutputStream(name string, file *RAMFile, checksum bool) *RAMOutputStream {
	crc := NewFakeHash32()
	if checksum {
		crc = crc32.NewIEEE()
	}

	stream := &RAMOutputStream{
		BaseIndexOutput: nil,
		bufferSize:      RAM_BUFFER_SIZE,
		file:            file,
		buffer:          new(bytes.Buffer),
		crc:             crc,
		idx:             0,
	}
	stream.BaseIndexOutput = NewBaseIndexOutput(name, stream)
	return stream
}

func (s *RAMOutputStream) Close() error {
	s.flush()
	s.file.setLength(int64(s.idx))
	return nil
}

func (s *RAMOutputStream) flush() {
	size := s.buffer.Len()

	dst := s.file.addBuffer(size)
	copy(dst, s.buffer.Bytes())
}

func (s *RAMOutputStream) Write(p []byte) (n int, err error) {
	if _, err := s.crc.Write(p); err != nil {
		return 0, err
	}

	s.idx += len(p)

	oldSize := s.buffer.Len()
	newSize := oldSize + len(p)
	if newSize < s.bufferSize {
		return s.buffer.Write(p)
	}

	if newSize == s.bufferSize {
		dst := s.file.addBuffer(s.bufferSize)
		copy(dst, s.buffer.Bytes())
		copy(dst[oldSize:], p)
	} else {
		dst := s.file.addBuffer(s.bufferSize)
		copy(dst, s.buffer.Bytes())
		copy(dst[oldSize:], p)
		s.buffer.Write(p[s.bufferSize-oldSize:])
	}

	return len(p), nil
}

func (s *RAMOutputStream) GetFilePointer() int64 {
	return int64(s.idx)
}

func (s *RAMOutputStream) GetChecksum() (uint32, error) {
	return s.crc.Sum32(), nil
}
