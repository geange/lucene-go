package index

import "github.com/geange/lucene-go/core/types"

// SortedNumericDocValues A list of per-document numeric values, sorted according to Long.CompareFn(long, long).
type SortedNumericDocValues interface {
	types.DocValuesIterator

	// NextValue Iterates to the next item in the current document. Do not call this more than
	// docValueCount times for the document.
	NextValue() (int64, error)

	// DocValueCount Retrieves the number of values for the current document. This must always be greater
	// than zero. It is illegal to call this method after advanceExact(int) returned false.
	DocValueCount() int
}

var _ DocValuesWriter = &SortedNumericDocValuesWriter{}

type SortedNumericDocValuesWriter struct {
}

func (s *SortedNumericDocValuesWriter) Flush(state *SegmentWriteState, sortMap DocMap, consumer DocValuesConsumer) error {
	//TODO implement me
	panic("implement me")
}

func (s *SortedNumericDocValuesWriter) GetDocValues() types.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}
