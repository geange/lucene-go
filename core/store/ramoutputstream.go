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
	return nil
}

func (s *RAMOutputStream) flush() {
	s.file.Write(s.buffer.Bytes())
}

func (s *RAMOutputStream) Write(p []byte) (n int, err error) {
	if _, err := s.crc.Write(p); err != nil {
		return 0, err
	}

	pSize := len(p)
	bSize := s.buffer.Len()
	if bSize+pSize < s.bufferSize {
		s.idx += pSize
		return s.buffer.Write(p)
	}

	wSize := s.bufferSize - bSize
	s.buffer.Write(p[:wSize])
	s.file.Write(s.buffer.Bytes())
	s.buffer.Reset()

	for start := wSize; wSize < pSize; start += s.bufferSize {
		end := start + s.bufferSize
		if end <= pSize {
			// 不用写缓存
			s.file.Write(p[start:end])
		} else {
			s.buffer.Write(p[start:])
			break
		}
	}
	s.idx += pSize
	return pSize, nil
}

func (s *RAMOutputStream) GetFilePointer() int64 {
	return int64(s.idx)
}

func (s *RAMOutputStream) GetChecksum() (uint32, error) {
	return s.crc.Sum32(), nil
}
