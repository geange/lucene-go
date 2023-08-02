package store

import (
	"slices"
	"sync/atomic"

	"github.com/geange/gods-generic/lists/arraylist"
)

type RAMFile struct {
	buffers    *arraylist.List[[]byte]
	length     *atomic.Int64
	directory  *RAMDirectory
	bufferSize int
}

func NewRAMFile(dir *RAMDirectory) *RAMFile {
	return &RAMFile{
		buffers:   arraylist.New[[]byte](),
		directory: dir,
	}
}

func (f *RAMFile) GetLength() int64 {
	return f.length.Load()
}

func (f *RAMFile) Clone() *RAMFile {
	file := &RAMFile{
		buffers:   arraylist.New[[]byte](),
		length:    &atomic.Int64{},
		directory: f.directory,
	}

	for _, v := range f.buffers.Values() {
		file.buffers.Add(slices.Clone(v))
	}
	return file
}

func (f *RAMFile) setLength(size int64) {
	f.length.Store(size)
}

func (f *RAMFile) addBuffer(size int) []byte {
	buf := make([]byte, size)
	f.buffers.Add(buf)
	return buf
}

func (f *RAMFile) getBuffer(index int) ([]byte, bool) {
	return f.buffers.Get(index)
}

func (f *RAMFile) numBuffers() int {
	return f.buffers.Size()
}
