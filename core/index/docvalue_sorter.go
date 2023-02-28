package index

import (
	"sort"
)

var _ sort.Interface = &DocValueSorter{}

type DocValueSorter struct {
	docs       []int
	comparator DocComparator
}

func NewDocValueSorter(docs []int, comparator DocComparator) *DocValueSorter {
	return &DocValueSorter{docs: docs, comparator: comparator}
}

func (d *DocValueSorter) Len() int {
	return len(d.docs)
}

func (d *DocValueSorter) Less(i, j int) bool {
	if d.comparator.Compare(d.docs[i], d.docs[j]) < 0 {
		return true
	}
	return false
}

func (d *DocValueSorter) Swap(i, j int) {
	d.docs[i], d.docs[j] = d.docs[j], d.docs[i]
}
