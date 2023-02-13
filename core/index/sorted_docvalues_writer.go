package index

var _ DocValuesWriter = &SortedDocValuesWriter{}

type SortedDocValuesWriter struct {
}

func (s *SortedDocValuesWriter) Flush(state *SegmentWriteState, sortMap DocMap, consumer DocValuesConsumer) error {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesWriter) GetDocValues() DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}
