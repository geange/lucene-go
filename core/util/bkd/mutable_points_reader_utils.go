package bkd

import (
	"bytes"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/packed"
	"sort"
)

var _ sort.Interface = &MutablePointValuesSorter{}

type MutablePointValuesSorter struct {
	config *Config
	maxDoc int
	reader types.MutablePointValues
	from   int
	to     int

	sortedByDocID bool
	prevDoc       int
	bitsPerDocId  int
}

func NewMutablePointValuesSorter(config *Config, maxDoc int, reader types.MutablePointValues,
	from, to int) *MutablePointValuesSorter {

	sortedByDocID := true
	prevDoc := 0

	for i := from; i < to; i++ {
		doc := reader.GetDocID(i)
		if doc < prevDoc {
			sortedByDocID = false
			break
		}
		prevDoc = doc
	}

	bitsPerDocId := 0
	if !sortedByDocID {
		bitsPerDocId, _ = packed.BitsRequired(int64(maxDoc - 1))
	}

	return &MutablePointValuesSorter{
		config:        config,
		maxDoc:        maxDoc,
		reader:        reader,
		from:          from,
		to:            to,
		sortedByDocID: sortedByDocID,
		prevDoc:       prevDoc,
		bitsPerDocId:  bitsPerDocId,
	}
}

func (m *MutablePointValuesSorter) Len() int {
	return m.to - m.from
}

func (m *MutablePointValuesSorter) Less(i, j int) bool {
	buf := new(bytes.Buffer)
	m.reader.GetValue(i, buf)
	bs1 := buf.Bytes()

	m.reader.GetValue(j, buf)
	bs2 := buf.Bytes()

	flag := bytes.Compare(bs1, bs2)

	if flag == 0 {
		if m.reader.GetDocID(i) < m.reader.GetDocID(j) {
			return true
		}
		return false
	}

	if flag < 0 {
		return true
	}
	return false
}

func (m *MutablePointValuesSorter) Swap(i, j int) {
	m.reader.Swap(m.from+i, m.from+j)
}

var _ sort.Interface = &IntroSorter{}

type IntroSorter struct {
	config              *Config
	sortedDim           int
	commonPrefixLengths []int
	reader              types.MutablePointValues
	from, to            int
	scratch1, scratch2  *bytes.Buffer
	start, dimEnd       int
}

func NewIntroSorter(config *Config, sortedDim int, commonPrefixLengths []int, reader types.MutablePointValues, from int, to int, scratch1 *bytes.Buffer, scratch2 *bytes.Buffer) *IntroSorter {
	start := sortedDim*config.bytesPerDim + commonPrefixLengths[sortedDim]
	dimEnd := sortedDim*config.bytesPerDim + config.bytesPerDim

	return &IntroSorter{
		config:              config,
		sortedDim:           sortedDim,
		commonPrefixLengths: commonPrefixLengths,
		reader:              reader,
		from:                from,
		to:                  to,
		scratch1:            scratch1,
		scratch2:            scratch2,
		start:               start,
		dimEnd:              dimEnd,
	}
}

func (r *IntroSorter) Len() int {
	return r.to - r.from
}

func (r *IntroSorter) Less(i, j int) bool {
	r.reader.GetValue(i, r.scratch1)
	r.reader.GetValue(j, r.scratch2)

	cmp := bytes.Compare(r.scratch1.Bytes()[r.start:r.dimEnd], r.scratch2.Bytes()[r.start:r.dimEnd])
	if cmp < 0 {
		return true
	}

	if cmp == 0 {
		fromIndex := r.config.packedIndexBytesLength
		toIndex := r.config.packedBytesLength
		cmp = bytes.Compare(r.scratch1.Bytes()[fromIndex:toIndex], r.scratch2.Bytes()[fromIndex:toIndex])
		if cmp < 0 {
			return true
		}

		if cmp == 0 {
			return r.reader.GetDocID(i) < r.reader.GetDocID(j)
		}
		return false
	}
	return false
}

func (r *IntroSorter) Swap(i, j int) {
	r.reader.Swap(i, j)
}

// Partition points around mid. All values on the left must be less than or equal to it and
// all values on the right must be greater than or equal to it.
func partition(config *Config, maxDoc, splitDim, commonPrefixLen int,
	reader types.MutablePointValues, from, to, mid int,
	scratch1, scratch2 *bytes.Buffer) {

	panic("")
}
