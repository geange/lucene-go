package bkd

import (
	"bytes"
	"sort"
)

// SortK 排序并保证index=K前面的值比K小，index从0开始
func SortK(data sort.Interface, k int) {
	begin, end := 0, data.Len()-1
	sortN(begin, end, k, data)
}

func sortN(from, to, k int, data sort.Interface) {
	for from < to {
		loc := sortPartition(from, to, data)
		if loc == k {
			return
		}
		if loc < k {
			sortN(loc+1, to, k, data)
			return
		}

		sortN(from, loc-1, k, data)
	}
}

func sortPartition(begin, end int, data sort.Interface) int {
	i, j := begin+1, end

	for i < j {
		if data.Less(begin, i) {
			data.Swap(i, j)
			j--
		} else {
			i++
		}
	}

	// 如果 values[begin] <= values[i]
	// !data.Less(begin, i) && !data.Less(i, begin) => values[begin] == values[i]
	if data.Less(begin, i) || (!data.Less(begin, i) && !data.Less(i, begin)) {
		i--
	}
	data.Swap(i, begin)
	return i
}

var _ sort.Interface = &heapRadixSort{}

type heapRadixSort struct {
	from, to    int
	dimOffset   int
	dimCmpBytes int
	dataOffset  int
	selector    *RadixSelector
	points      *HeapPointWriter
}

func (h *heapRadixSort) Len() int {
	return h.to - h.from
}

func (h *heapRadixSort) Less(i, j int) bool {
	i += h.from
	j += h.from

	aFromIndex := i*h.selector.config.BytesPerDoc() + h.dimOffset
	bFromIndex := i*h.selector.config.BytesPerDoc() + h.dimOffset

	cmp := bytes.Compare(
		h.points.block[aFromIndex:aFromIndex+h.dimCmpBytes],
		h.points.block[bFromIndex:bFromIndex+h.dimCmpBytes],
	)
	if cmp == 0 {
		// 比较数据，data bytes
		aFromIndex = i*h.selector.config.BytesPerDoc() + h.dataOffset
		bFromIndex = i*h.selector.config.BytesPerDoc() + h.dataOffset

		cmp = bytes.Compare(
			h.points.block[aFromIndex:aFromIndex+h.dimCmpBytes],
			h.points.block[bFromIndex:bFromIndex+h.dimCmpBytes],
		)
		return cmp < 0
	}

	return cmp < 0
}

func (h *heapRadixSort) Swap(i, j int) {
	i += h.from
	j += h.from
	h.points.Swap(i, j)
}
