package memory

import "github.com/geange/lucene-go/core/util"

type SliceByteStartArray struct {
	*util.DirectBytesStartArray

	start []int // the start offset in the IntBlockPool per term
	end   []int // the end pointer in the IntBlockPool for the postings slice per term
	freq  []int // the term frequency
}

func NewSliceByteStartArray(initSize int) *SliceByteStartArray {
	return &SliceByteStartArray{
		DirectBytesStartArray: util.NewDirectBytesStartArray(initSize),
		start:                 nil,
		end:                   nil,
		freq:                  nil,
	}
}

func (s *SliceByteStartArray) Init() []int {
	ord := s.DirectBytesStartArray.Init()
	size := len(ord)

	size = util.Oversize(size, 4)

	s.start = make([]int, size)
	s.end = make([]int, size)
	s.freq = make([]int, size)
	return ord
}

func (s *SliceByteStartArray) Grow() []int {
	ord := s.DirectBytesStartArray.Grow()
	size := len(ord)
	if len(s.start) < size {
		s.start = util.Grow(s.start, size)
		s.end = util.Grow(s.end, size)
		s.freq = util.Grow(s.freq, size)
	}
	return ord
}

func (s *SliceByteStartArray) Clear() []int {
	s.start = nil
	s.end = nil

	return s.DirectBytesStartArray.Clear()
}
