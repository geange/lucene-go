package index

var _ DocValuesWriter = &SortedSetDocValuesWriter{}

type SortedSetDocValuesWriter struct {
}

func (s *SortedSetDocValuesWriter) Flush(state *SegmentWriteState, sortMap DocMap, consumer DocValuesConsumer) error {
	//TODO implement me
	panic("implement me")
}

func (s *SortedSetDocValuesWriter) GetDocValues() DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}
