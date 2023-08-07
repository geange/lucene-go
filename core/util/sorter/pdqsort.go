package sorter

import (
	"sort"
)

const (
	// SINGLE_MEDIAN_THRESHOLD
	// Below this size threshold, the partition selection is simplified to a single median.
	SINGLE_MEDIAN_THRESHOLD = 40
)

type pdqSorter struct {
	data sort.Interface
}

func NewPdqSorter(data sort.Interface) Sorter {
	return &pdqSorter{
		data: data,
	}
}

func (s *pdqSorter) Sort(from, to int) {
	sort.Sort(&pdqSorterRange{
		from: from,
		to:   to,
		data: s.data,
	})
}

var _ sort.Interface = &pdqSorterRange{}

type pdqSorterRange struct {
	from int
	to   int
	data sort.Interface
}

func (s *pdqSorterRange) Len() int {
	return s.to - s.from
}

func (s *pdqSorterRange) Less(i, j int) bool {
	return s.data.Less(s.from+i, s.from+j)
}

func (s *pdqSorterRange) Swap(i, j int) {
	s.data.Swap(s.from+i, s.from+j)
}
