package memory

import (
	"github.com/geange/lucene-go/core/util/array"
)

type sliceByteStartArray struct {
	initSize   int
	bytesStart []int
	start      []int // the start offset in the IntBlockPool per term
	end        []int // the end pointer in the IntBlockPool for the postings slice per term
	freq       []int // the term frequency
}

func newSliceByteStartArray(initSize int) *sliceByteStartArray {
	return &sliceByteStartArray{
		initSize: initSize,
	}
}

func (s *sliceByteStartArray) Init() []int {
	s.bytesStart = make([]int, s.initSize)
	ord := s.bytesStart
	size := len(ord)

	size = array.Oversize(size, 4)

	s.start = make([]int, size)
	s.end = make([]int, size)
	s.freq = make([]int, size)
	return ord
}

func (s *sliceByteStartArray) Grow() []int {
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

func (s *sliceByteStartArray) Clear() []int {
	s.start = s.start[:0]
	s.end = s.end[:0]
	s.freq = s.freq[:0]
	s.bytesStart = s.bytesStart[:0]
	return s.bytesStart
}
