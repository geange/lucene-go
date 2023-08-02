package memory

import (
	"github.com/geange/lucene-go/core/util/array"
)

type sliceByteStartArray struct {
	initSize   int
	bytesStart []uint32
	start      []int // the start offset in the IntBlockPool per term
	end        []int // the end pointer in the IntBlockPool for the postings slice per term
	freq       []int // the term frequency
}

func newSliceByteStartArray(initSize int) *sliceByteStartArray {
	return &sliceByteStartArray{
		initSize: initSize,
	}
}

func (s *sliceByteStartArray) Init() []uint32 {
	s.bytesStart = make([]uint32, s.initSize)
	ord := s.bytesStart
	size := len(ord)

	size = array.Oversize(size, 4)

	s.start = make([]int, size)
	s.end = make([]int, size)
	s.freq = make([]int, size)
	return ord
}

func (s *sliceByteStartArray) Grow() []uint32 {
	s.bytesStart = append(s.bytesStart, 0)
	ord := s.bytesStart

	size := len(ord)
	if len(s.start) < size {
		s.start = array.Grow(s.start, size)
		s.end = array.Grow(s.end, size)
		s.freq = array.Grow(s.freq, size)
	}
	return ord
}

func (s *sliceByteStartArray) Clear() []uint32 {
	s.start = s.start[:0]
	s.end = s.end[:0]
	s.freq = s.freq[:0]
	s.bytesStart = s.bytesStart[:0]
	return s.bytesStart
}
