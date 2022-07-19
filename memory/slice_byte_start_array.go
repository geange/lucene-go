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
