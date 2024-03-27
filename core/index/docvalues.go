package index

import (
	"sort"

	"github.com/geange/lucene-go/core/types"
)

type DocValues struct {
}

func GetNumeric(reader LeafReader, field string) (NumericDocValues, error) {
	dv, err := reader.GetNumericDocValues(field)
	if err != nil {
		return nil, err
	}
	return dv, nil
}

func GetSorted(reader LeafReader, field string) (SortedDocValues, error) {
	dv, err := reader.GetSortedDocValues(field)
	if err != nil {
		return nil, err
	}
	if dv != nil {
		return dv, nil
	}

	panic("")
}

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

type DocValuesWriter interface {
	Flush(state *SegmentWriteState, sortMap DocMap, consumer DocValuesConsumer) error
	GetDocValues() types.DocIdSetIterator
}

// IsCacheable
// Returns true if the specified docvalues fields have not been updated
func IsCacheable(ctx LeafReaderContext, fields ...string) bool {
	for _, field := range fields {
		fi := ctx.LeafReader().GetFieldInfos().FieldInfo(field)
		if fi != nil && fi.GetDocValuesGen() > -1 {
			return false
		}
	}
	return true
}
