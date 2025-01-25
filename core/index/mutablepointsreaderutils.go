package index

import (
	"bytes"
	"sort"

	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/bkd"
	"github.com/geange/lucene-go/core/util/packed"
	"github.com/geange/lucene-go/core/util/radixselector"
)

// SortByDim
// Sort points on the given dimension.
func SortByDim(config *bkd.Config, sortedDim int, commonPrefixLengths []int,
	reader types.MutablePointValues, from, to int,
	scratch1, scratch2 *bytes.Buffer) {

	start := sortedDim*config.BytesPerDoc() + commonPrefixLengths[sortedDim]
	dimEnd := sortedDim*config.BytesPerDoc() + config.BytesPerDoc()
	// No need for a fancy radix sort here, this is called on the leaves only so
	// there are not many values to sort
	sorter := &innerSortByDim{
		from:   from,
		to:     to,
		start:  start,
		dimEnd: dimEnd,
		reader: reader,
		buf1:   scratch1,
		buf2:   scratch2,
		config: config,
	}
	sort.Sort(sorter)
}

var _ sort.Interface = &innerSortByDim{}

type innerSortByDim struct {
	from   int
	to     int
	start  int
	dimEnd int
	reader types.MutablePointValues
	buf1   *bytes.Buffer
	buf2   *bytes.Buffer
	config *bkd.Config
}

func (r *innerSortByDim) Len() int {
	return r.from - r.to + 1
}

func (r *innerSortByDim) Less(i, j int) bool {
	r.buf1.Reset()
	r.buf2.Reset()

	r.reader.GetValue(i, r.buf1)
	r.reader.GetValue(j, r.buf2)

	bs1 := r.buf1.Bytes()
	bs2 := r.buf2.Bytes()

	from, to := r.start, r.dimEnd
	cmp := bytes.Compare(bs1[from:to], bs2[from:to])
	if cmp == 0 {
		from, to = r.config.PackedIndexBytesLength(), r.config.PackedBytesLength()
		cmp = bytes.Compare(bs1[from:to], bs2[from:to])
		if cmp == 0 {
			cmp = r.reader.GetDocID(i) - r.reader.GetDocID(j)
		}
	}
	if cmp < 0 {
		return true
	}
	return false
}

func (r *innerSortByDim) Swap(i, j int) {
	r.reader.Swap(i, j)
}

// Partition points around mid. All values on the left must be less than or equal to it and all values on the right must be greater than or equal to it.
func Partition(config *bkd.Config, maxDoc, splitDim, commonPrefixLen int,
	reader types.MutablePointValues, from, to, mid int,
	_scratch1, _scratch2 *bytes.Buffer) {

	dimOffset := splitDim*config.BytesPerDoc() + commonPrefixLen
	dimCmpBytes := config.BytesPerDoc() - commonPrefixLen
	dataCmpBytes := (config.NumDims()-config.NumIndexDims())*config.BytesPerDoc() + dimCmpBytes
	bitsPerDocId, _ := packed.BitsRequired(int64(maxDoc - 1))

	radix := radixselector.NewRadixSelector(&radixselector.RadixSelectorConfig{
		MaxLength: dataCmpBytes + (bitsPerDocId+7)/8,
		FnByteAt: func(i, k int) int {
			if k < dimCmpBytes {
				return int(reader.GetByteAt(i, dimOffset+k))
			} else if k < dataCmpBytes {
				return int(reader.GetByteAt(i, config.PackedIndexBytesLength()+k-dimCmpBytes))
			} else {
				shift := bitsPerDocId - ((k - dataCmpBytes + 1) << 3)
				return reader.GetDocID(i) >> max(0, shift)
			}
		},
		FnSwap: func(i, j int) {
			reader.Swap(i, j)
		},
		FnGetFallbackSelector: func(k int) util.Selector {
			panic("")
		},
	})
	radix.Select(from, to, mid)
}

type innerSelector struct {
	pivot    *bytes.Buffer
	pivotDoc int
	reader   types.MutablePointValues
	*util.IntroSelector
}

func newInnerSelector() *innerSelector {
	inner := &innerSelector{
		pivot:         nil,
		pivotDoc:      0,
		reader:        nil,
		IntroSelector: nil,
	}
	inner.IntroSelector = util.NewIntroSelector(&util.IntroSelectorConfig{
		FnSwap:         inner.Swap,
		FnSetPivot:     nil,
		FnComparePivot: nil,
		FnCompare:      nil,
	})

	return inner
}

func (r *innerSelector) Swap(i, j int) {
	r.reader.Swap(i, j)
}

func (r *innerSelector) SetPivot(i int) {
	r.reader.GetValue(i, r.pivot)
	r.pivotDoc = r.reader.GetDocID(i)
}

func (r *innerSelector) comparePivot(j int) {

}
