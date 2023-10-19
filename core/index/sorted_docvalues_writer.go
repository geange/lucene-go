package index

import "github.com/geange/lucene-go/core/types"

var _ DocValuesWriter = &SortedDocValuesWriter{}

type SortedDocValuesWriter struct {
}

func (s *SortedDocValuesWriter) Flush(state *SegmentWriteState, sortMap DocMap, consumer DocValuesConsumer) error {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesWriter) GetDocValues() types.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}
