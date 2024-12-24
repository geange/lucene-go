package index

import (
	"sort"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

type DocValues struct {
}

func GetNumeric(reader index.LeafReader, field string) (index.NumericDocValues, error) {
	dv, err := reader.GetNumericDocValues(field)
	if err != nil {
		return nil, err
	}
	return dv, nil
}

func GetSorted(reader index.LeafReader, field string) (index.SortedDocValues, error) {
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
	comparator index.DocComparator
}

func NewDocValueSorter(docs []int, comparator index.DocComparator) *DocValueSorter {
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
	Flush(state *index.SegmentWriteState, sortMap index.DocMap, consumer index.DocValuesConsumer) error
	GetDocValues() types.DocIdSetIterator
}

// IsCacheable
// Returns true if the specified docvalues fields have not been updated
func IsCacheable(ctx index.LeafReaderContext, fields ...string) bool {
	for _, field := range fields {
		fi := ctx.LeafReader().GetFieldInfos().FieldInfo(field)
		if fi != nil && fi.GetDocValuesGen() > -1 {
			return false
		}
	}
	return true
}
