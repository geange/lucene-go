//go:build linux || darwin

package store

import (
	"io"

	"golang.org/x/exp/mmap"
)

var (
	_ DataInput = &MmapDataInput{}
)

type MmapDataInput struct {
	*BaseDataInput

	pos    int64
	reader *mmap.ReaderAt
	isEOF  bool
}

func NewMmapDataInput(file string) (*MmapDataInput, error) {
	reader, err := mmap.Open(file)
	if err != nil {
		return nil, err
	}

	input := &MmapDataInput{
		pos:    0,
		reader: reader,
		isEOF:  false,
	}
	input.BaseDataInput = NewBaseDataInput(input)
	return input, nil
}

func (m *MmapDataInput) Read(p []byte) (n int, err error) {
	if m.isEOF {
		return 0, io.EOF
	}

	less := m.reader.Len() - int(m.pos)

	if len(p) > less {
		size, err := m.reader.ReadAt(p[:less], m.pos)
		if err != nil {
			return 0, err
		}
		m.pos += int64(size)
		m.isEOF = true
		return size, nil
	}

	size, err := m.reader.ReadAt(p, m.pos)
	if err != nil {
		return 0, err
	}
	m.pos += int64(size)
	return size, nil
}

func (m *MmapDataInput) Clone() CloneReader {
	input := &MmapDataInput{
		pos:    m.pos,
		reader: m.reader,
		isEOF:  m.isEOF,
	}
	input.BaseDataInput = NewBaseDataInput(input)
	return input
}

func (m *MmapDataInput) Close() error {
	return m.reader.Close()
}
