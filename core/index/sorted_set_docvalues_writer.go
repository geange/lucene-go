package index

import "github.com/geange/lucene-go/core/types"

var _ DocValuesWriter = &SortedSetDocValuesWriter{}

type SortedSetDocValuesWriter struct {
}

func (s *SortedSetDocValuesWriter) Flush(state *SegmentWriteState, sortMap DocMap, consumer DocValuesConsumer) error {
	//TODO implement me
	panic("implement me")
}

func (s *SortedSetDocValuesWriter) GetDocValues() types.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}
