package store

import (
	"iter"
	"slices"
	"sync/atomic"
)

type RAMFile struct {
	buffers   [][]byte
	size      *atomic.Int64
	directory *RAMDirectory
}

func NewRAMFile(dir *RAMDirectory) *RAMFile {
	return &RAMFile{
		buffers:   make([][]byte, 0),
		size:      new(atomic.Int64),
		directory: dir,
	}
}

func (f *RAMFile) GetLength() int64 {
	return f.size.Load()
}

func (f *RAMFile) Clone() *RAMFile {
	dst := &RAMFile{
		buffers:   make([][]byte, 0),
		size:      &atomic.Int64{},
		directory: f.directory,
	}

	for _, buf := range f.buffers {
		dst.buffers = append(dst.buffers, slices.Clone(buf))
	}
	dst.size.Store(f.size.Load())
	return dst
}

func (f *RAMFile) Write(p []byte) {
	if len(p) == 0 {
		return
	}
	buf := slices.Clone(p)
	f.buffers = append(f.buffers, buf)
	f.size.Add(int64(len(p)))
}

func (f *RAMFile) GetBuffer(n int) ([]byte, bool) {
	if n >= len(f.buffers) || n < 0 {
		return nil, false
	}
	return f.buffers[n], true
}

func (f *RAMFile) NumBuffers() int {
	return len(f.buffers)
}

func (f *RAMFile) Iterator() iter.Seq[byte] {
	return func(yield func(byte) bool) {
		for _, buffer := range f.buffers {
			for _, b := range buffer {
				if !yield(b) {
					return
				}
			}
		}
	}
}
